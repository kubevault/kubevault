package apis

import (
	"github.com/go-openapi/spec"
	core "k8s.io/api/core/v1"
	"k8s.io/kube-openapi/pkg/common"
)

var (
	EnableStatusSubresource bool
)

func SetNameSchema(openapiSpec map[string]common.OpenAPIDefinition) {
	// ref: https://github.com/kubedb/project/issues/166
	// https://github.com/kubernetes/apimachinery/blob/94ebb086c69b9fec4ddbfb6a1433d28ecca9292b/pkg/util/validation/validation.go#L153
	var maxLength int64 = 63
	openapiSpec["k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"].Schema.SchemaProps.Properties["name"] = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
			Type:        []string{"string"},
			Format:      "",
			Pattern:     `^[a-z]([-a-z0-9]*[a-z0-9])?$`,
			MaxLength:   &maxLength,
		},
	}
}

const (
	// Specifies the path where auth is enabled
	AuthPathKey = "kubevault.com/auth-path"

	// required fields:
	// - Secret.Data["token"] - a vault token
	SecretTypeTokenAuth core.SecretType = "kubevault.com/token"

	// required for SecretTypeTokenAut
	TokenAuthTokenKey = "token"

	// required fields:
	// - Secret.Data["access_key_id"] - aws access key id
	// - Secret.Data["secret_access_key"] - aws access secret key
	//
	// optional fields:
	// - Secret.Annotations["kubevault.com/aws.header-value"] - specifies the header value that required if X-Vault-AWS-IAM-Server-ID Header is set
	// - Secret.Annotations["kubevault.com/auth-path"] - Specifies the path where aws auth is enabled
	SecretTypeAWSAuth core.SecretType = "kubevault.com/aws"

	// required for SecretTypeAWSAuth
	AWSAuthAccessKeyIDKey = "access_key_id"
	// required for SecretTypeAWSAuth
	AWSAuthAccessSecretKey = "secret_access_key"
	// optional for SecretTypeAWSAuth
	AWSAuthSecurityTokenKey = "security_token"

	// Specifies the header value that required if X-Vault-AWS-IAM-Server-ID Header is set
	// optional for annotation for  SecretTypeAWSAuth
	AWSHeaderValueKey = "kubevault.com/aws.header-value"
)
