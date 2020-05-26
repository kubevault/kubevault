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

package v1alpha1

import (
	"fmt"

	"kubevault.dev/operator/api/crds"

	"kmodules.xyz/client-go/apiextensions"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/clusterid"
)

func (_ VaultPolicy) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceVaultPolicies))
}

func (v VaultPolicy) GetKey() string {
	return ResourceVaultPolicy + "/" + v.Namespace + "/" + v.Name
}

func (v VaultPolicy) PolicyName() string {
	if v.Spec.VaultPolicyName != "" {
		return v.Spec.VaultPolicyName
	}

	cluster := "-"
	if clusterid.ClusterName() != "" {
		cluster = clusterid.ClusterName()
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, v.Namespace, v.Name)
}

func (v VaultPolicy) OffshootSelectors() map[string]string {
	return map[string]string{
		"app":          "vault",
		"vault_policy": v.Name,
	}
}

func (v VaultPolicy) OffshootLabels() map[string]string {
	return meta_util.FilterKeys("kubevault.com", v.OffshootSelectors(), v.Labels)
}

func (v VaultPolicy) IsValid() error {
	return nil
}
