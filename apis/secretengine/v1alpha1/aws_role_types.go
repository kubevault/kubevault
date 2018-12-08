package v1alpha1

import (
	"github.com/appscode/go/encoding/json/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	ResourceKindAWSRole = "AWSRole"
	ResourceAWSRole     = "awsrole"
	ResourceAWSRoles    = "awsroles"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AWSRole
type AWSRole struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AWSRoleSpec   `json:"spec,omitempty"`
	Status            AWSRoleStatus `json:"status,omitempty"`
}

type AWSCredentialType string

const (
	AWSCredentialIAMUser         AWSCredentialType = "iam_user"
	AWSCredentialAssumedRole     AWSCredentialType = "assumed_role"
	AWSCredentialFederationToken AWSCredentialType = "federation_token"
)

// AWSRoleSpec contains connection information, AWS role info, etc
type AWSRoleSpec struct {
	AuthManagerRef *appcat.AppReference `json:"authManagerRef,omitempty"`

	Config *AWSConfig `json:"config"`

	// links:
	// 	- https://www.vaultproject.io/api/secret/aws/index.html

	// Specifies the type of credential to be used when retrieving credentials from the role
	CredentialType AWSCredentialType `json:"credentialType"`

	// Specifies the ARNs of the AWS roles this Vault role is allowed to assume.
	// Required when credential_type is assumed_role and prohibited otherwise
	RoleARNs []string `json:"roleARNs,omitempty"`

	// Specifies the ARNs of the AWS managed policies to be attached to IAM users when they are requested.
	// Valid only when credential_type is iam_user. When credential_type is iam_user,
	// at least one of policy_arns or policy_document must be specified.
	PolicyARNs []string `json:"policyARNs,omitempty"`

	// The IAM policy document for the role. The behavior depends on the credential type.
	// With iam_user, the policy document will be attached to the IAM user generated and
	// augment the permissions the IAM user has. With assumed_role and federation_token,
	// the policy document will act as a filter on what the credentials can do.
	PolicyDocument string `json:"policyDocument,omitempty"`

	// The default TTL for STS credentials. When a TTL is not specified when STS credentials are requested,
	// and a default TTL is specified on the role, then this default TTL will be used.
	// Valid only when credential_type is one of assumed_role or federation_token
	DefaultSTSTTL string `json:"defaultSTSTTL,omitempty"`

	// The max allowed TTL for STS credentials (credentials TTL are capped to max_sts_ttl).
	// Valid only when credential_type is one of assumed_role or federation_token
	MaxSTSTTL string `json:"maxSTSTTL,omitempty"`

	// Specifies the IAM policy in JSON format.
	Policy string `json:"policy,omitempty"`

	// Specifies the full ARN reference to the desired existing policy.
	ARN string `json:"arn,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AWSRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of AWSRole objects
	Items []AWSRole `json:"items,omitempty"`
}

const (
	AWSCredentialAccessKeyKey = "access_key"
	AWSCredentialSecretKeyKey = "secret_key"
)

// https://www.vaultproject.io/api/secret/aws/index.html#configure-root-iam-credentials
// AWSConfig contains information to communicate with AWS
type AWSConfig struct {
	// Specifies the secret containing AWS access key ID and secret access key
	// secret.Data:
	//	- access_key=<value>
	//  - secret_key=<value>
	CredentialSecret string `json:"credentialSecret"`

	// Specifies the AWS region
	Region string `json:"region"`

	// Specifies a custom HTTP IAM enminidpoint to use
	IAMEndpoint string `json:"iamEndpoint,omitempty"`

	//Specifies a custom HTTP STS endpoint to use
	STSEndpoint string `json:"stsEndpoint,omitempty"`

	// Number of max retries the client should use for recoverable errors.
	// The default (-1) falls back to the AWS SDK's default behavior
	MaxRetries *int `json:"maxRetries,omitempty"`

	LeaseConfig *LeaseConfig `json:"leaseConfig,omitempty"`
}

// https://www.vaultproject.io/api/secret/aws/index.html#configure-lease
// LeaseConfig contains lease configuration
type LeaseConfig struct {
	// Specifies the lease value provided as a string duration with time suffix.
	// "h" (hour) is the largest suffix.
	Lease string `json:"lease"`

	// Specifies the maximum lease value provided as a string duration with time suffix.
	// "h" (hour) is the largest suffix
	LeaseMax string `json:"leaseMax"`
}

type AWSRolePhase string

type AWSRoleStatus struct {
	Phase AWSRolePhase `json:"phase,omitempty"`

	// observedGeneration is the most recent generation observed for this AWSRole. It corresponds to the
	// AWSRole's generation, which is updated on mutation by the API Server.
	ObservedGeneration *types.IntHash `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a AWSRole current state.
	Conditions []AWSRoleCondition `json:"conditions,omitempty"`
}

// AWSRoleCondition describes the state of a AWSRole at a certain point.
type AWSRoleCondition struct {
	// Type of AWSRole condition.
	Type string `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}
