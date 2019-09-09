package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	config "kubevault.dev/operator/apis/config/v1alpha1"
	api "kubevault.dev/operator/apis/engine/v1alpha1"
)

// GCP secret engine policies
const SecretEnginePolicyGCP = `
path "{{ . }}/config" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/roleset/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/token/*" {
	capabilities = ["create", "update", "read"]
}

path "{{ . }}/key/*" {
	capabilities = ["create", "update", "read"]
}
`

// AWS secret engine policies
const SecretEnginePolicyAWS = `
path "{{ . }}/config/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/roles/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/creds/*" {
	capabilities = ["create", "update", "read"]
}

path "{{ . }}/sts/*" {
	capabilities = ["create", "update", "read"]
}
`

// Azure secret engine policies
const SecretEnginePolicyAzure = `
path "{{ . }}/config" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/roles/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/creds/*" {
	capabilities = ["create", "update", "read"]
}
`

// Database secret engine policies
const SecretEnginePolicyDatabase = `
path "{{ . }}/config/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/roles/*" {
	capabilities = ["create", "update", "read", "delete"]
}

path "{{ . }}/creds/*" {
	capabilities = ["create", "update", "read"]
}
`

type KubernetesAuthRole struct {
	Data RoleData `json:"data"`
}
type RoleData struct {
	BoundServiceAccountNames      []string    `json:"bound_service_account_names"`
	BoundServiceAccountNamespaces []string    `json:"bound_service_account_namespaces"`
	TokenTtl                      json.Number `json:"token_ttl"`
	TokenMaxTtl                   json.Number `json:"token_max_ttl"`
	TokenPolicies                 []string    `json:"token_policies"`
	TokenBoundCidrs               []string    `json:"token_bound_cidrs"`
	TokenExplicitMaxTtl           json.Number `json:"token_explicit_max_ttl"`
	TokenNoDefaultPolicy          bool        `json:"token_no_default_policy"`
	TokenNumUses                  json.Number `json:"token_num_uses"`
	TokenPeriod                   json.Number `json:"token_period"`
	TokenType                     string      `json:"token_type"`
}

func (seClient *SecretEngine) CreatePolicy() error {
	var policy bytes.Buffer
	var policyTemplate string
	engSpec := seClient.secretEngine.Spec

	if engSpec.GCP != nil {
		policyTemplate = SecretEnginePolicyGCP
	} else if engSpec.AWS != nil {
		policyTemplate = SecretEnginePolicyAWS
	} else if engSpec.Azure != nil {
		policyTemplate = SecretEnginePolicyAzure
	} else if engSpec.MySQL != nil || engSpec.MongoDB != nil || engSpec.Postgres != nil {
		policyTemplate = SecretEnginePolicyDatabase
	} else {
		return errors.New("unknown secret engine type")
	}

	tpl := template.Must(template.New("").Parse(policyTemplate))
	err := tpl.Execute(&policy, seClient.path)
	if err != nil {
		return errors.Wrap(err, "failed to execute policy template")
	}

	policyName := seClient.secretEngine.GetPolicyName()
	err = seClient.vaultClient.Sys().PutPolicy(policyName, policy.String())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create vault policy: %s", policyName))
	}

	return nil
}

func GetPolicyControllerRoleInfo(appClient appcat_cs.AppcatalogV1alpha1Interface, vClient *vaultapi.Client, secretEngine *api.SecretEngine) (*KubernetesAuthRole, string, error) {
	// Get appbinding referred in SecretEngine spec
	vApp, err := appClient.AppBindings(secretEngine.Namespace).Get(secretEngine.Spec.VaultRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get appbinding for secret engine")
	}

	// Currently secret engine feature works only with kubernetes auth method
	if vApp.Spec.Parameters == nil || vApp.Spec.Parameters.Raw == nil {
		return nil, "", errors.New("appbinding parameters is nil")
	}

	var cf config.VaultServerConfiguration
	err = json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to unmarshal appbinding parameters")
	}
	if cf.PolicyControllerRole == "" {
		return nil, "", errors.New("policyControllerRole is empty")
	}

	// Get policy controller role data from vault
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", cf.PolicyControllerRole)
	req := vClient.NewRequest("GET", path)
	resp, err := vClient.RawRequest(req)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed making GET request to vault")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var role KubernetesAuthRole
	err = json.Unmarshal(bodyBytes, &role)
	if err != nil {
		return nil, "", err
	}

	return &role, cf.PolicyControllerRole, nil
}

func (seClient *SecretEngine) UpdateAuthRole() error {
	// Get policy controller role name from appbinding and
	// get role data from vault
	role, roleName, err := GetPolicyControllerRoleInfo(seClient.appClient, seClient.vaultClient, seClient.secretEngine)
	if err != nil {
		return errors.Wrap(err, "failed to get policy controller role information")
	}

	// Check whether the policy already exist or not
	exist := false
	policyName := seClient.secretEngine.GetPolicyName()
	for _, value := range role.Data.TokenPolicies {
		if policyName == value {
			exist = true
			break
		}
	}

	// if not exist append the policy to the slice
	if !exist {
		role.Data.TokenPolicies = append(role.Data.TokenPolicies, policyName)
	}

	// update the policy controller role with new policies
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", roleName)
	req := seClient.vaultClient.NewRequest("POST", path)
	err = req.SetJSONBody(role.Data)
	if err != nil {
		return err
	}

	_, err = seClient.vaultClient.RawRequest(req)
	return err
}

func (seClient *SecretEngine) DeletePolicyAndUpdateRole() error {
	// delete policy created for this secret engine
	policyName := seClient.secretEngine.GetPolicyName()
	err := seClient.vaultClient.Sys().DeletePolicy(policyName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to delete vault policy: %s", policyName))
	}

	// get policy controller role name from appbinding and
	// also get policy controller role data from vault
	role, roleName, err := GetPolicyControllerRoleInfo(seClient.appClient, seClient.vaultClient, seClient.secretEngine)
	if err != nil {
		return errors.Wrap(err, "failed to get policy controller role information")
	}

	// get the location the policy if exist
	exist := false
	var index int
	for id, value := range role.Data.TokenPolicies {
		if policyName == value {
			exist = true
			index = id
			break
		}
	}

	// if the policy exist in TokenPolices
	// delete it from the list
	if exist {
		// swap the value at `index` at the end of the slice
		role.Data.TokenPolicies[len(role.Data.TokenPolicies)-1], role.Data.TokenPolicies[index] = role.Data.TokenPolicies[index], role.Data.TokenPolicies[len(role.Data.TokenPolicies)-1]
		// reduce slice size by one
		role.Data.TokenPolicies = role.Data.TokenPolicies[:len(role.Data.TokenPolicies)-1]
	}

	// Update role with new policies
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", roleName)
	req := seClient.vaultClient.NewRequest("POST", path)
	err = req.SetJSONBody(role.Data)
	if err != nil {
		return err
	}

	_, err = seClient.vaultClient.RawRequest(req)
	return err
}
