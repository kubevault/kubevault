package aws

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/engine/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type AWSRole struct {
	awsRole     *api.AWSRole
	vaultClient *vaultapi.Client
	kubeClient  kubernetes.Interface
	awsPath     string // Specifies the path where aws is enabled
}

// checks whether aws is enabled or not
func (a *AWSRole) IsAWSEnabled() (bool, error) {
	mnt, err := a.vaultClient.Sys().ListMounts()
	if err != nil {
		return false, errors.Wrap(err, "failed to list mounted secrets engines")
	}

	mntPath := a.awsPath + "/"
	for k := range mnt {
		if k == mntPath {
			return true, nil
		}
	}
	return false, nil
}

// EnableDatabase enables aws secret engine
// It first checks whether aws is enabled or not
func (a *AWSRole) EnableAWS() error {
	enabled, err := a.IsAWSEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return nil
	}

	err = a.vaultClient.Sys().Mount(a.awsPath, &vaultapi.MountInput{
		Type: "aws",
	})
	if err != nil {
		return err
	}
	return nil
}

// https://www.vaultproject.io/api/secret/aws/index.html#configure-root-iam-credentials
func (a *AWSRole) CreateConfig() error {
	if a.vaultClient == nil {
		return errors.New("vault client is nil")
	}
	if a.awsRole == nil {
		return errors.New("AWSRole is nil")
	}
	if a.awsPath == "" {
		return errors.New("aws engine path is empty")
	}

	path := fmt.Sprintf("/v1/%s/config/root", a.awsPath)
	req := a.vaultClient.NewRequest("POST", path)

	payload := map[string]interface{}{}
	cfg := a.awsRole.Spec.Config
	if cfg == nil {
		return errors.New("aws secret engine config is nil")
	}
	if cfg.MaxRetries != nil {
		payload["max_retries"] = *cfg.MaxRetries
	}
	if cfg.Region != "" {
		payload["region"] = cfg.Region
	}
	if cfg.IAMEndpoint != "" {
		payload["iam_endpoint"] = cfg.IAMEndpoint
	}
	if cfg.STSEndpoint != "" {
		payload["sts_endpoint"] = cfg.STSEndpoint
	}

	if cfg.CredentialSecret != "" {
		sr, err := a.kubeClient.CoreV1().Secrets(a.awsRole.Namespace).Get(cfg.CredentialSecret, metav1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to get aws credential secret")
		}

		if val, ok := sr.Data[api.AWSCredentialAccessKeyKey]; ok {
			payload["access_key"] = string(val)
		}
		if val, ok := sr.Data[api.AWSCredentialSecretKeyKey]; ok {
			payload["secret_key"] = string(val)
		}
	}

	if err := req.SetJSONBody(payload); err != nil {
		return errors.Wrap(err, "failed to load payload in config create request")
	}

	_, err := a.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrap(err, "failed to create aws config")
	}

	// set lease config
	if cfg.LeaseConfig != nil {
		path := fmt.Sprintf("/v1/%s/config/lease", a.awsPath)
		req := a.vaultClient.NewRequest("POST", path)

		payload := map[string]interface{}{
			"lease":     cfg.LeaseConfig.Lease,
			"lease_max": cfg.LeaseConfig.LeaseMax,
		}
		if err := req.SetJSONBody(payload); err != nil {
			return errors.Wrap(err, "failed to load payload in create lease config request")
		}

		_, err := a.vaultClient.RawRequest(req)
		if err != nil {
			return errors.Wrap(err, "failed to create aws lease config")
		}
	}
	return nil
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
	}
	if roleSpec.DefaultSTSTTL != "" {
		payload["default_sts_ttl"] = roleSpec.DefaultSTSTTL
	}
	if roleSpec.MaxSTSTTL != "" {
		payload["max_sts_ttl"] = roleSpec.MaxSTSTTL
	}
	if roleSpec.Policy != "" {
		payload["policy"] = roleSpec.Policy
	}
	if roleSpec.ARN != "" {
		payload["arn"] = roleSpec.ARN
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
		return errors.Wrapf(err, "failed to delete database role %s", name)
	}
	return nil
}
