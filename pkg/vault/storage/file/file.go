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

package file

import (
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	VaultFileSystemVolumeName = "vault-filesystem-backend"
)

var fileStorageFmt = `
storage "file" {
%s
}
`

type Options struct {
	api.FileSpec
	claimName string
}

func NewOptions(kubeClient kubernetes.Interface, vaultServer *api.VaultServer, s *api.FileSpec) (*Options, error) {
	if s == nil {
		return nil, errors.New("fileSpec is empty")
	}
	if vaultServer == nil {
		return nil, errors.New("vaultServer object is empty")
	}

	// Set the pvc name and labels if given.
	// Otherwise default to VaultServer's name, namespace and labels.
	objMeta := v1.ObjectMeta{
		Name:      vaultServer.Name,
		Namespace: vaultServer.Namespace,
		Labels:    vaultServer.OffshootLabels(),
	}

	if s.VolumeClaimTemplate.Name != "" {
		objMeta.Name = s.VolumeClaimTemplate.Name
	}

	if s.VolumeClaimTemplate.Labels != nil {
		objMeta.Labels = s.VolumeClaimTemplate.Labels
	}

	// Create or Patch the requested PVC
	_, _, err := core_util.CreateOrPatchPVC(kubeClient, objMeta, func(claim *core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
		// pvc.spec is immutable except spec.resources.request field.
		// But values need to be set while creating the pvc for the first time.
		// Here, "Spec.AccessModes" will be "nil" in two cases; invalid pvc template
		// & creating pvc for the first time.
		if claim.Spec.AccessModes == nil {
			claim.Spec = s.VolumeClaimTemplate.Spec
		}

		// Update labels
		claim.Labels = objMeta.Labels

		// Update the only mutable field.
		claim.Spec.Resources.Requests = s.VolumeClaimTemplate.Spec.Resources.Requests
		return claim
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pvc %s/%s", objMeta.Namespace, objMeta.Name)
	}

	return &Options{
		*s,
		objMeta.Name,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.Path == "" || o.claimName == "" {
		return errors.New("path or pvc name is empty")
	}

	pt.Spec.Volumes = append(pt.Spec.Volumes, core.Volume{
		Name: VaultFileSystemVolumeName,
		VolumeSource: core.VolumeSource{
			PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
				ClaimName: o.claimName,
			},
		},
	})

	pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, core.VolumeMount{
		Name:      VaultFileSystemVolumeName,
		MountPath: o.Path,
	})

	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
// GetGcsConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Path != "" {
		params = append(params, fmt.Sprintf(`path = "%s"`, o.Path))
	}

	storageCfg := fmt.Sprintf(fileStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
