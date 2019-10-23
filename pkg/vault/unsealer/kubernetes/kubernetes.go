package kubernetes

import (
	"fmt"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
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
}

func NewOptions(s api.KubernetesSecretSpec) (*Options, error) {
	return &Options{
		s,
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

	cont.Args = append(cont.Args, args...)
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
