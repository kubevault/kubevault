package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindUserPassAuthConfiguration = "UserPassAuthConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserPassAuthConfiguration defines a Vault UserPass auth configuration.
type UserPassAuthConfiguration struct {
	metav1.TypeMeta `json:",inline,omitempty"`

	// Specifies the path where userpass auth is enabled
	// default : userpass
	AuthPath string `json:"authPath,omitempty"`
}
