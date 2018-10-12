package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindKubernetesAuthConfiguration = "KubernetesAuthConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubernetesAuthConfiguration defines a Vault Kuberenetes auth configuration.
// https://www.vaultproject.io/api/auth/kubernetes/index.html#login
type KubernetesAuthConfiguration struct {
	metav1.TypeMeta `json:",inline,omitempty"`

	// Name of the role against which the login is being attempted.
	Role string `json:"role"`
}
