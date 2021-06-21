/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"fmt"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	ModeKubernetesSecret = "kubernetes-secret"
)

type Options struct {
	api.KubernetesSecretSpec
	Backend string
}

func NewOptions(s api.KubernetesSecretSpec, backend api.VaultServerBackend) (*Options, error) {
	return &Options{
		KubernetesSecretSpec: s,
		Backend:              string(backend),
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	var args []string
	var cont core.Container

	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultUnsealerContainerName {
			cont = c
		}
	}

	args = append(args, fmt.Sprintf("--mode=%s", ModeKubernetesSecret))

	if o.SecretName != "" {
		args = append(args, fmt.Sprintf("--k8s.secret-name=%s", o.SecretName))
	}
	args = append(args, fmt.Sprintf("--storage-backend=%s", o.Backend))

	cont.Args = append(cont.Args, args...)

	var envs []core.EnvVar
	envs = append(envs, core.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &core.EnvVarSource{
			FieldRef: &core.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	})
	cont.Env = core_util.UpsertEnvVars(cont.Env, envs...)

	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(prefix, namespace string) []rbac.Role {
	var roles []rbac.Role

	role := rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prefix + "-unsealer-secret-reader",
			Namespace: namespace,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{core.GroupName},
				Resources: []string{"secrets"},
				Verbs:     []string{"create", "get", "patch"},
			},
		},
	}

	roles = append(roles, role)
	return roles
}
