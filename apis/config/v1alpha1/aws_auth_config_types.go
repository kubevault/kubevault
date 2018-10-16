package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindAWSAuthConfiguration = "AWSAuthConfiguration"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSAuthConfiguration defines a Vault AWS auth configuration.
// https://www.vaultproject.io/api/auth/aws/index.html#login
type AWSAuthConfiguration struct {
	metav1.TypeMeta `json:",inline,omitempty"`

	// Name of the role against which the login is being attempted.
	// If role is not specified, then the login endpoint looks for a
	// role bearing the name of the AMI ID of the EC2 instance that
	// is trying to login if using the ec2 auth method, or the
	// "friendly name" (i.e., role name or username) of the IAM
	// principal authenticated. If a matching role is not found,
	// login fails.
	Role string `json:"role,omitempty"`

	// Specifies the header value that required if X-Vault-AWS-IAM-Server-ID Header is set
	HeaderValue string `json:"headerValue,omitempty"`

	// Specifies the path where aws auth is enabled
	// default : aws
	AuthPath string `json:"authPath,omitempty"`
}
