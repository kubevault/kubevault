package aws

import (
	"encoding/json"
	"fmt"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

type AWSRole struct {
	awsRole     *api.AWSRole
	vaultClient *vaultapi.Client
	kubeClient  kubernetes.Interface
	awsPath     string // Specifies the path where aws is enabled
}

// https://www.vaultproject.io/api/secret/aws/index.html#create-update-role
func (a *AWSRole) CreateRole() error {
	if a.vaultClient == nil {
		return errors.New("vault client is nil")
	}
	if a.awsRole == nil {
		return errors.New("AWSRole is nil")
	}
	if a.awsPath == "" {
		return errors.New("aws engine path is empty")
	}

	path := fmt.Sprintf("/v1/%s/roles/%s", a.awsPath, a.awsRole.RoleName())
	req := a.vaultClient.NewRequest("POST", path)

	roleSpec := a.awsRole.Spec
	payload := map[string]interface{}{
		"credential_type": roleSpec.CredentialType,
	}
	if len(roleSpec.RoleARNs) > 0 {
		payload["role_arns"] = roleSpec.RoleARNs
	}
	if len(roleSpec.PolicyARNs) > 0 {
		payload["policy_arns"] = roleSpec.PolicyARNs
	}
	if roleSpec.PolicyDocument != "" {
		payload["policy_document"] = roleSpec.PolicyDocument
	} else if roleSpec.Policy != nil {
		doc, err := json.Marshal(roleSpec.Policy)
		if err != nil {
			return fmt.Errorf("failed to serialize spec.policy of AWSRole %s/%s. Reason: %v", a.awsRole.Namespace, a.awsRole.Name, err)
		}
		payload["policy_document"] = string(doc)
	}

	if roleSpec.DefaultSTSTTL != "" {
		payload["default_sts_ttl"] = roleSpec.DefaultSTSTTL
	}
	if roleSpec.MaxSTSTTL != "" {
		payload["max_sts_ttl"] = roleSpec.MaxSTSTTL
	}

	if err := req.SetJSONBody(payload); err != nil {
		return errors.Wrap(err, "failed to load payload in aws create role request")
	}

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create aws role")
	}
	return nil
}

// https://www.vaultproject.io/api/secret/aws/index.html#delete-role
//
// DeleteRole deletes role
// It's safe to call multiple time. It doesn't give
// error even if respective role doesn't exist
func (a *AWSRole) DeleteRole(name string) error {
	path := fmt.Sprintf("/v1/%s/roles/%s", a.awsPath, name)
	req := a.vaultClient.NewRequest("DELETE", path)

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to delete aws role %s", name)
	}
	return nil
}
