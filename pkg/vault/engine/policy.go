package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	config "kubevault.dev/operator/apis/config/v1alpha1"
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

func (secretEngineClient *SecretEngine) CreatePolicy() error {
	var policy bytes.Buffer
	var policyTemplate string
	engSpec := secretEngineClient.secretEngine.Spec

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
	err := tpl.Execute(&policy, secretEngineClient.path)
	if err != nil {
		return errors.Wrap(err, "failed to execute policy template")
	}

	policyName := secretEngineClient.secretEngine.GetPolicyName()
	err = secretEngineClient.vaultClient.Sys().PutPolicy(policyName, policy.String())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create vault policy: %s", policyName))
	}

	return nil
}

func (secretEngineClient *SecretEngine) UpdateAuthRole() error {

	vApp, err := secretEngineClient.appClient.AppBindings(secretEngineClient.secretEngine.Namespace).Get(secretEngineClient.secretEngine.Spec.VaultRef.Name, metav1.GetOptions{})
	if err != nil {
		errors.Wrap(err, "failed to get appbinding for secret engine")
	}

	// Currently secret engine feature works only with kubernetes auth method
	if vApp.Spec.Parameters == nil || vApp.Spec.Parameters.Raw == nil {
		return errors.New("appbinding parameters is nil")
	}

	var cf config.VaultServerConfiguration
	err = json.Unmarshal(vApp.Spec.Parameters.Raw, &cf)
	if err != nil {
		errors.Wrap(err, "failed to unmarshal appbinding parameters")
	}
	if cf.PolicyControllerRole == "" {
		errors.New("policyControllerRole is empty")
	}
	path := fmt.Sprintf("/v1/auth/kubernetes/role/%s", cf.PolicyControllerRole)
	req := secretEngineClient.vaultClient.NewRequest("GET", path)
	resp, err := secretEngineClient.vaultClient.RawRequest(req)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var role KubernetesAuthRole
	err = json.Unmarshal(bodyBytes, &role)
	if err != nil {
		return err
	}
	exist := false
	policyName := secretEngineClient.secretEngine.GetPolicyName()
	for _, value := range role.Data.TokenPolicies {
		if policyName == value {
			exist = true
			break
		}
	}
	if !exist {
		role.Data.TokenPolicies = append(role.Data.TokenPolicies, policyName)
	}

	req = secretEngineClient.vaultClient.NewRequest("POST", path)
	err = req.SetJSONBody(role.Data)
	if err != nil {
		return err
	}

	_, err = secretEngineClient.vaultClient.RawRequest(req)
	return err
}
