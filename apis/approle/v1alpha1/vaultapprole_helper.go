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

	apiextensions "kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/tools/clusterid"
)

func (v VaultAppRole) GetKey() string {
	return ResourceVaultAppRole + "/" + v.Namespace + "/" + v.Name
}

func (_ VaultAppRole) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceVaultAppRoles))
}

func (v VaultAppRole) AppRoleName() string {
	if v.Spec.RoleName != "" {
		return v.Spec.RoleName
	}

	cluster := "-"
	if clusterid.ClusterName() != "" {
		cluster = clusterid.ClusterName()
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, v.Namespace, v.Name)
}

func (v VaultAppRole) GeneratePayLoad() (map[string]interface{}, error) {

	ret := map[string]interface{}{
		"role_name":               v.AppRoleName(),
		"bind_secret_id":          v.Spec.BindSecretID,
		"secret_id_bound_cidrs":   v.Spec.SecretIDBoundCidrs,
		"secret_id_num_uses":      v.Spec.SecretIDNumUses,
		"secret_id_ttl":           v.Spec.SecretIDTTL,
		"enable_local_secret_ids": v.Spec.EnableLocalSecretIDs,
		"token_ttl":               v.Spec.TokenTTL,
		"token_max_ttl":           v.Spec.TokenMaxTTL,
		"token_policies":          v.Spec.TokenPolicies,
		"token_bound_cidrs":       v.Spec.TokenBoundCidrs,
		"token_explicit_max_ttl":  v.Spec.TokenExplicitMaxTTL,
		"token_no_default_policy": v.Spec.TokenNoDefaultPolicy,
		"token_num_uses":          v.Spec.TokenNumUses,
		"token_period":            v.Spec.TokenPeriod,
		"token_type":              v.Spec.TokenType,
	}

	return ret, nil
}
