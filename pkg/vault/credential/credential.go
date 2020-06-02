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

package credential

import (
	"context"
	"encoding/json"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core_util "kmodules.xyz/client-go/core/v1"
	rbac_util "kmodules.xyz/client-go/rbac/v1"
)

type CredManager struct {
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	secretEngine SecretEngine
}

// Creates a kubernetes secret containing database credential
func (c *CredManager) CreateSecret(name string, namespace string, credSecret *vaultapi.Secret) error {
	data := map[string][]byte{}
	if credSecret != nil {
		var err error
		data, err = c.secretEngine.ParseCredential(credSecret)
		if err != nil {
			return errors.Wrap(err, "failed to parse credential secret")
		}
	}

	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), c.kubeClient, obj, func(in *core.Secret) *core.Secret {
		in.Data = data
		core_util.EnsureOwnerReference(in, c.secretEngine.GetOwnerReference())
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create/update secret %s/%s", namespace, name)
	}
	return nil
}

// Creates kubernetes role
func (c *CredManager) CreateRole(name string, namespace string, secretName string) error {
	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	_, _, err := rbac_util.CreateOrPatchRole(context.TODO(), c.kubeClient, obj, func(in *rbac.Role) *rbac.Role {
		in.Rules = []rbac.PolicyRule{
			{
				APIGroups: []string{
					"", // represents core api
				},
				Resources: []string{
					"secrets",
				},
				ResourceNames: []string{
					secretName,
				},
				Verbs: []string{
					"get",
				},
			},
		}

		core_util.EnsureOwnerReference(in, c.secretEngine.GetOwnerReference())

		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create rbac role %s/%s", namespace, name)
	}
	return nil
}

// Create kubernetes role binding
func (c *CredManager) CreateRoleBinding(name string, namespace string, roleName string, subjects []rbac.Subject) error {
	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	_, _, err := rbac_util.CreateOrPatchRoleBinding(context.TODO(), c.kubeClient, obj, func(in *rbac.RoleBinding) *rbac.RoleBinding {
		in.RoleRef = rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     roleName,
		}
		in.Subjects = subjects

		core_util.EnsureOwnerReference(in, c.secretEngine.GetOwnerReference())
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create/update rbac role binding %s/%s", namespace, name)
	}
	return nil
}

// https://www.vaultproject.io/api/system/leases.html#read-lease
//
// Whether or not lease is expired in vault
// In vault, lease is revoked if lease is expired
func (c *CredManager) IsLeaseExpired(leaseID string) (bool, error) {
	if leaseID == "" {
		return true, nil
	}

	req := c.vaultClient.NewRequest("PUT", "/v1/sys/leases/lookup")
	err := req.SetJSONBody(map[string]string{
		"lease_id": leaseID,
	})
	if err != nil {
		return false, errors.WithStack(err)
	}

	resp, err := c.vaultClient.RawRequest(req)
	if resp == nil && err != nil {
		return false, errors.WithStack(err)
	}

	defer resp.Body.Close()
	errResp := vaultapi.ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if len(errResp.Errors) > 0 {
		return true, nil
	}
	return false, nil
}

// RevokeLease revokes respective lease
// It's safe to call multiple time. It doesn't give
// error even if respective lease_id doesn't exist
// but it will give an error if lease_id is empty
func (c *CredManager) RevokeLease(leaseID string) error {
	err := c.vaultClient.Sys().Revoke(leaseID)
	if err != nil {
		return errors.Wrap(err, "failed to revoke lease")
	}
	return nil
}

// Gets credential from vault
func (c *CredManager) GetCredential() (*vaultapi.Secret, error) {
	return c.secretEngine.GetSecret()
}
