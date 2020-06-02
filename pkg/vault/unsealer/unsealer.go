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

package unsealer

import (
	"fmt"
	"time"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	sa_util "kubevault.dev/operator/pkg/util"
	"kubevault.dev/operator/pkg/vault/unsealer/aws"
	"kubevault.dev/operator/pkg/vault/unsealer/azure"
	"kubevault.dev/operator/pkg/vault/unsealer/google"
	k8s "kubevault.dev/operator/pkg/vault/unsealer/kubernetes"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/analytics"
	"kmodules.xyz/client-go/tools/cli"
	"kmodules.xyz/client-go/tools/clientcmd"
)

const (
	K8sTokenReviewerJwtEnv = "K8S_TOKEN_REVIEWER_JWT"
	timeout                = 30 * time.Second
	timeInterval           = 2 * time.Second
)

type Unsealer interface {
	Apply(pt *core.PodTemplateSpec) error
	GetRBAC(prefix, namespace string) []rbac.Role
}

type unsealerSrv struct {
	restConfig *rest.Config
	kc         kubernetes.Interface
	vs         *api.VaultServer
	unsealer   Unsealer
	image      string
}

func newUnsealer(s *api.UnsealerSpec) (Unsealer, error) {
	if s.Mode.KubernetesSecret != nil {
		return k8s.NewOptions(*s.Mode.KubernetesSecret)
	} else if s.Mode.GoogleKmsGcs != nil {
		return google.NewOptions(*s.Mode.GoogleKmsGcs)
	} else if s.Mode.AwsKmsSsm != nil {
		return aws.NewOptions(*s.Mode.AwsKmsSsm)
	} else if s.Mode.AzureKeyVault != nil {
		return azure.NewOptions(*s.Mode.AzureKeyVault)
	} else {
		return nil, errors.New("unsealer mode is not valid/defined")
	}
}

func NewUnsealerService(restConfig *rest.Config, vs *api.VaultServer, image string) (Unsealer, error) {
	if vs == nil {
		return nil, errors.New("VaultServer is nil")
	}
	if vs.Spec.Unsealer == nil {
		glog.Infoln(".spec.unsealer is nil")
		return nil, nil
	}

	unslr, err := newUnsealer(vs.Spec.Unsealer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unsealer service")
	}

	kc, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes client")
	}
	return &unsealerSrv{
		restConfig: restConfig,
		vs:         vs,
		kc:         kc,
		image:      image,
		unsealer:   unslr,
	}, nil
}

// Apply will do:
// 	- add unsealer container for vault
//	- add additional env, volume mounts etc if required
func (u *unsealerSrv) Apply(pt *core.PodTemplateSpec) error {
	if u == nil {
		return nil
	}

	var args []string

	unslr := u.vs.Spec.Unsealer
	cont := core.Container{
		Name:  util.VaultUnsealerContainerName,
		Image: u.image,
	}
	args = append(args,
		"run",
		"--v=3",
	)

	if unslr.SecretShares != 0 {
		args = append(args, fmt.Sprintf("--secret-shares=%d", unslr.SecretShares))
	}
	if unslr.SecretThreshold != 0 {
		args = append(args, fmt.Sprintf("--secret-threshold=%d", unslr.SecretThreshold))
	}

	if unslr.RetryPeriodSeconds != 0 {
		p := time.Second * unslr.RetryPeriodSeconds
		args = append(args, fmt.Sprintf("--retry-period=%s", p.String()))
	}
	if unslr.OverwriteExisting {
		args = append(args, "--overwrite-existing=true")
	}

	if u.vs.Spec.TLS != nil && u.vs.Spec.TLS.CABundle != nil {
		args = append(args, fmt.Sprintf("--vault.ca-cert=%s", u.vs.Spec.TLS.CABundle))
	}

	// Add kubernetes auth flags
	args = append(args, fmt.Sprintf("--auth.k8s-host=%s", u.restConfig.Host))

	err := rest.LoadTLSFiles(u.restConfig)
	if err != nil {
		return errors.Wrap(err, "failed to load TLS files from rest config for kubernetes auth")
	}
	args = append(args, fmt.Sprintf("--auth.k8s-ca-cert=%s", u.restConfig.CAData))

	// Add env for kubernetes auth
	secret, err := sa_util.TryGetJwtTokenSecretNameFromServiceAccount(u.kc, u.vs.ServiceAccountForTokenReviewer(), u.vs.Namespace, timeInterval, timeout)
	if err != nil {
		return errors.Wrapf(err, "failed to get jwt token secret of service account(%s/%s)", u.vs.Namespace, u.vs.ServiceAccountForTokenReviewer())
	}
	cont.Env = core_util.UpsertEnvVars(cont.Env, core.EnvVar{
		Name: K8sTokenReviewerJwtEnv,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: secret.Name,
				},
				Key: core.ServiceAccountTokenKey,
			},
		},
	}, core.EnvVar{
		Name:  analytics.Key,
		Value: cli.AnalyticsClientID,
	})

	// Add flags for policy
	args = append(args, fmt.Sprintf("--policy-manager.name=%s", u.vs.PolicyNameForPolicyController()))
	args = append(args, fmt.Sprintf("--policy-manager.service-account-name=%s", u.vs.ServiceAccountName()))
	args = append(args, fmt.Sprintf("--policy-manager.service-account-namespace=%s", u.vs.Namespace))
	args = append(args, fmt.Sprintf("--use-kubeapiserver-fqdn-for-aks=%v", clientcmd.UseKubeAPIServerFQDNForAKS()))

	cont.Args = append(cont.Args, args...)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	err = u.unsealer.Apply(pt)
	if err != nil {
		return err
	}
	return nil
}

// GetRBAC return rbac roles required by unsealer
func (u *unsealerSrv) GetRBAC(prefix, namespace string) []rbac.Role {
	if u == nil {
		return nil
	}
	return u.unsealer.GetRBAC(prefix, namespace)
}
