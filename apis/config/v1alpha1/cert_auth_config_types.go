package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindCertAuthConfiguration = "CertAuthConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertAuthConfiguration defines a Vault Cert auth configuration.
// https://www.vaultproject.io/api/auth/cert/index.html#login-with-tls-certificate-method
type CertAuthConfiguration struct {
	metav1.TypeMeta `json:",inline,omitempty"`

	// Authenticate against only the named certificate role,
	// returning its policy list if successful. If not set,
	// defaults to trying all certificate roles and returning
	// any one that matches.
	Name string `json:"name,omitempty"`
}
