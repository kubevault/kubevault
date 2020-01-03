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
	"errors"
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	core "k8s.io/api/core/v1"
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

func NewOptions(kubeClient kubernetes.Interface, namespace string, s *api.FileSpec) (*Options, error) {
	if s == nil {
		return nil, errors.New("fileSpec is empty")
	}

	var claimName string
	if s.VolumeClaimTemplate != nil && s.VolumeClaimTemplate.Name != "" {
		// Generate PVC object out of VolumeClainTemplate
		pvc := s.VolumeClaimTemplate.ToCorePVC()

		// Set pvc's namespace to vaultServer's namespace, if not provided
		if pvc.Namespace == "" {
			pvc.Namespace = namespace
		}

		// Create or Patch the requested PVC
		_, _, err := core_util.CreateOrPatchPVC(kubeClient, pvc.ObjectMeta, func(claim *core.PersistentVolumeClaim) *core.PersistentVolumeClaim {
			claim = pvc
			return claim
		})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to create pvc %s/%s", pvc.Namespace, pvc.Name))
		}
		claimName = pvc.Name
	}
	return &Options{
		*s,
		claimName,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.claimName != "" {
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
	}
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
