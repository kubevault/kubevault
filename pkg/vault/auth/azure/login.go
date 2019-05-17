package azure

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/kubevault/operator/apis"
	config "github.com/kubevault/operator/apis/config/v1alpha1"
	"github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/auth/types"
	vaultutil "github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

// ref:
// -https://www.vaultproject.io/api/auth/azure/index.html#sample-payload-2

// Fields that are required to authenticate the request
type auth struct {
	vClient           *vaultapi.Client
	role              string
	path              string
	signedJwt         string
	subscriptionId    string
	resourceGroupName string
	vmName            string
	vmssName          string
}

// ref:
// - https://www.vaultproject.io/api/auth/azure/index.html
// - https://www.vaultproject.io/docs/auth/azure.html

func New(vApp *appcat.AppBinding, secret *corev1.Secret) (*auth, error) {
	if vApp.Spec.Parameters == nil {
		return nil, errors.New("parameters are not provided in AppBinding spec")
	}

	var cf config.VaultServerConfiguration
	err := json.Unmarshal([]byte(vApp.Spec.Parameters.Raw), &cf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal parameters")
	}

	if cf.PolicyControllerRole == "" {
		return nil, errors.New("PolicyControllerRole is missing")
	}

	cfg, err := vaultutil.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client from config")
	}

	signedJwt, ok := secret.Data[apis.AzureMSIToken]
	if !ok {
		return nil, errors.Errorf("msiToken is missing in %s/%s", secret.Namespace, secret.Name)
	}

	authPath := string(v1alpha1.AuthTypeAzure)
	if val, ok := secret.Annotations[apis.AuthPathKey]; ok && len(val) > 0 {
		authPath = val
	}

	subscriptionId := ""
	if val, ok := secret.Annotations[apis.AzureSubscriptionId]; ok && len(val) > 0 {
		subscriptionId = val
	}

	resourceGroupName := ""
	if val, ok := secret.Annotations[apis.AzureResourceGroupName]; ok && len(val) > 0 {
		resourceGroupName = val
	}

	vmName := ""
	if val, ok := secret.Annotations[apis.AzureVmName]; ok && len(val) > 0 {
		vmName = val
	}

	vmssName := ""
	if val, ok := secret.Annotations[apis.AzureVmssName]; ok && len(val) > 0 {
		vmssName = val
	}

	return &auth{
		vClient:           vc,
		role:              cf.PolicyControllerRole,
		path:              authPath,
		signedJwt:         string(signedJwt),
		subscriptionId:    subscriptionId,
		resourceGroupName: resourceGroupName,
		vmName:            vmName,
		vmssName:          vmssName,
	}, nil
}

// Login will log into vault and return client token
func (a *auth) Login() (string, error) {
	path := fmt.Sprintf("/v1/auth/%s/login", a.path)
	req := a.vClient.NewRequest("POST", path)

	payload := make(map[string]interface{})
	payload["role"] = a.role
	payload["jwt"] = a.signedJwt

	if len(a.subscriptionId) > 0 {
		payload["subscription_id"] = a.subscriptionId
	}
	if len(a.resourceGroupName) > 0 {
		payload["resource_group_name"] = a.resourceGroupName
	}
	if len(a.vmName) > 0 {
		payload["vm_name"] = a.vmName
	}
	if len(a.vmssName) > 0 {
		payload["vmss_name"] = a.vmssName
	}

	if err := req.SetJSONBody(payload); err != nil {
		return "", err
	}

	resp, err := a.vClient.RawRequest(req)
	if err != nil {
		return "", err
	}

	var loginResp types.AuthLoginResponse
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}
	return loginResp.Auth.ClientToken, nil
}
