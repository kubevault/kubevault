package kubernetes

import (
	"fmt"

	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (o *Options) Apply(pt *corev1.PodTemplateSpec, cont *corev1.Container) error {
	var args []string

	args = append(args, fmt.Sprintf("--mode=%s", ModeKubernetesSecret))

	if o.SecretName != "" {
		args = append(args, fmt.Sprintf("--k8s.secret-name=%s", o.SecretName))
	}

	cont.Args = append(cont.Args, args...)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(namespace string) []rbac.Role {
	var roles []rbac.Role

	role := rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault-unsealer-kubernetes-secret-access",
			Namespace: namespace,
		},
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secrets"},
				Verbs:     []string{"create", "get", "update", "patch"},
			},
		},
	}

	roles = append(roles, role)

	return roles
}

func (o *Options) GetSecrets(namespace string) ([]corev1.Secret, error) {
	return nil, nil
}
