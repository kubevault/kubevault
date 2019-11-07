/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package consul

import (
	"fmt"
	"path/filepath"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// TLS related file name for consul
	ConsulTLSAssetDir    = "/etc/vault/storage/consul/tls/"
	ConsulClientCaName   = "ca.crt"
	ConsulClientCertName = "client.crt"
	ConsulClientKeyName  = "client.key"
)

var consulStorageFmt = `
storage "consul" {
%s
}
`

type Options struct {
	kc        kubernetes.Interface
	namespace string
	api.ConsulSpec
}

func NewOptions(kc kubernetes.Interface, namespace string, s api.ConsulSpec) (*Options, error) {
	return &Options{
		kc,
		namespace,
		s,
	}, nil
}

// Apply will do:
// - If TLSSecretName is provided, then add volume for consul tls
// - If ACLTokenSecretName is provided, then set environment variable
func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	consulTLSAssetVolume := "vault-consul-tls"
	if o.TLSSecretName != "" {
		// mount tls secret
		pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
			Name: consulTLSAssetVolume,
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: o.TLSSecretName,
				},
			},
		})

		pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      consulTLSAssetVolume,
			MountPath: ConsulTLSAssetDir,
			ReadOnly:  true,
		})
	}
	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/consul.html
//
// Note:
// - Secret `TLSSecretName` mounted in `ConsulTLSAssetDir`
// - Secret `ACLTokenSecretName` will be used to aclToken from secret
//
// GetStorageConfig creates consul storage config from ConsulSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Address != "" {
		params = append(params, fmt.Sprintf(`address = "%s"`, o.Address))
	}
	if o.CheckTimeout != "" {
		params = append(params, fmt.Sprintf(`check_timeout = "%s"`, o.CheckTimeout))
	}
	if o.ConsistencyMode != "" {
		params = append(params, fmt.Sprintf(`consistency_mode = "%s"`, o.ConsistencyMode))
	}
	if o.DisableRegistration != "" {
		params = append(params, fmt.Sprintf(`disable_registration = "%s"`, o.DisableRegistration))
	}
	if o.MaxParallel != "" {
		params = append(params, fmt.Sprintf(`max_parallel = "%s"`, o.MaxParallel))
	}
	if o.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, o.Path))
	}
	if o.Scheme != "" {
		params = append(params, fmt.Sprintf(`scheme = "%s"`, o.Scheme))
	}
	if o.Service != "" {
		params = append(params, fmt.Sprintf(`service = "%s"`, o.Service))
	}
	if o.ServiceTags != "" {
		params = append(params, fmt.Sprintf(`service_tags = "%s"`, o.ServiceTags))
	}
	if o.ServiceAddress != "" {
		params = append(params, fmt.Sprintf(`service_address = "%s"`, o.ServiceAddress))
	}
	// Get ALC token from secret
	if o.ACLTokenSecretName != "" {
		secret, err := o.kc.CoreV1().Secrets(o.namespace).Get(o.ACLTokenSecretName, metav1.GetOptions{})
		if err != nil {
			return "", errors.Wrapf(err, "failed to get secret %s/%s", o.namespace, o.ACLTokenSecretName)

		}
		if value, exist := secret.Data["aclToken"]; !exist {
			return "", errors.Wrapf(err, "Data field is empty in %s/%s", o.namespace, o.ACLTokenSecretName)
		} else {
			params = append(params, fmt.Sprintf(`token = "%s"`, string(value)))
		}

	}
	if o.SessionTTL != "" {
		params = append(params, fmt.Sprintf(`session_ttl = "%s"`, o.SessionTTL))
	}
	if o.LockWaitTime != "" {
		params = append(params, fmt.Sprintf(`lock_wait_time = "%s"`, o.LockWaitTime))
	}
	if o.TLSSecretName != "" {
		params = append(params, fmt.Sprintf(`tls_ca_file = "%s"`, filepath.Join(ConsulTLSAssetDir, ConsulClientCaName)),
			fmt.Sprintf(`tls_cert_file = "%s"`, filepath.Join(ConsulTLSAssetDir, ConsulClientCertName)),
			fmt.Sprintf(`tls_key_file = "%s"`, filepath.Join(ConsulTLSAssetDir, ConsulClientKeyName)))
	}
	if o.TLSMinVersion != "" {
		params = append(params, fmt.Sprintf(`tls_min_version = "%s"`, o.TLSMinVersion))
	}
	if o.TLSSkipVerify {
		params = append(params, fmt.Sprintf(`tls_skip_verify = true`))
	}

	storageCfg := fmt.Sprintf(consulStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
