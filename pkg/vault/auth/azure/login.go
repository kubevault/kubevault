/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"encoding/json"
	"fmt"

	"kubevault.dev/operator/apis"
	"kubevault.dev/operator/apis/kubevault/v1alpha1"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"
	vaultutil "kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
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

func New(authInfo *authtype.AuthInfo) (*auth, error) {
	if authInfo == nil {
		return nil, errors.New("authentication information is empty")
	}
	if authInfo.VaultApp == nil {
		return nil, errors.New("AppBinding is empty")
	}

	vApp := authInfo.VaultApp
	cfg, err := vaultutil.VaultConfigFromAppBinding(vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault config from AppBinding")
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault client from config")
	}

	if authInfo.Secret == nil {
		return nil, errors.New("authentication secret is missing")
	}

	secret := authInfo.Secret
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

	if authInfo.VaultRole == "" {
		return nil, errors.New("Vault role is empty")
	}

	return &auth{
		vClient:           vc,
		role:              authInfo.VaultRole,
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

	var loginResp authtype.AuthLoginResponse
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}
	return loginResp.Auth.ClientToken, nil
}
