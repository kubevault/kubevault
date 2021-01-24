/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"unicode"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

////
// strip whitespace from the left until the margin ('|') or a non-space character is hit.
//
func stripMargin(s string) string {
	var stripped strings.Builder
	margin := '|'

	for _, line := range strings.Split(s, "\n") {
		hitMargin := false
		strippedLine := strings.TrimLeftFunc(line, func(c rune) bool {
			if c == margin {
				hitMargin = true
				return true
			}

			if unicode.IsSpace(c) {
				return !hitMargin
			}

			return false
		})

		stripped.WriteString(strippedLine)
		stripped.WriteString("\n")
	}

	return stripped.String()
}

var expectedPolicies = map[string]string{
	"gcpPolicyTest1": stripMargin(`
		|path "gcp/config" {
		|	capabilities = ["create", "update", "read", "delete"]
		|}

		|path "gcp/roleset/*" {
		|	capabilities = ["create", "update", "read", "delete"]
		|}

		|path "gcp/token/*" {
		|	capabilities = ["create", "update", "read"]
		|}

		|path "gcp/key/*" {
		|	capabilities = ["create", "update", "read"]
		|}

		|path "/sys/leases/*" {
		|	capabilities = ["create","update"]
		|}`),

	"gcpPolicyTest2": stripMargin(`
		|path "my-gcp-path/config" {
		|	capabilities = ["create", "update", "read", "delete"]
		|}

		|path "my-gcp-path/roleset/*" {
		|	capabilities = ["create", "update", "read", "delete"]
		|}

		|path "my-gcp-path/token/*" {
		|	capabilities = ["create", "update", "read"]
		|}

		|path "my-gcp-path/key/*" {
		|	capabilities = ["create", "update", "read"]
		|}

		|path "/sys/leases/*" {
		|	capabilities = ["create","update"]
		|}`),
	"kvv2": stripMargin(`
    |path "secret/config" {
    |	capabilities = ["create", "update", "read", "delete"]
    }`),

	"kvv2-custom-path": stripMargin(`
    |path "kvv2/config" {
    |	capabilities = ["create", "update", "read", "delete"]
    }`),
}

const aclBasePath = "/v1/sys/policies/acl"

func NewFakeVaultPolicyServer(requestRecorder *map[string]int) *httptest.Server {
	router := mux.NewRouter()

	recordPath := func(path string) {
		if requestRecorder != nil {
			(*requestRecorder)[path] = (*requestRecorder)[path] + 1
		}
	}

	router.HandleFunc(fmt.Sprintf("%s/{path}", aclBasePath), func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := vars["path"]
		body := r.Body
		data, _ := ioutil.ReadAll(body)
		var newdata map[string]interface{}
		_ = json.Unmarshal(data, &newdata)

		recordPath(fmt.Sprintf("/v1/sys/policies/acl/%s", path))

		fail := func(message string) {
			fail(message, w)
		}

		success := func() {
			success(w)
		}

		var expectedPolicyName, expectedPolicy, policy string
		var rawPolicy interface{}
		var ok bool

		if expectedPolicyName = r.Header.Get(KVTestHeaderExpectedPolicy); len(expectedPolicyName) == 0 {
			fail("No expected policy name. Unable to verify policy")
			return
		}

		if expectedPolicy == "none" {
			fail(fmt.Sprintf("Expected no policy to be configured, but got policy named '%s' with content:\n%v", path, newdata))
			return
		}

		if expectedPolicy, ok = expectedPolicies[expectedPolicyName]; !ok {
			fail(fmt.Sprintf("Unknown expected policy: %s", expectedPolicyName))
			return
		}

		if rawPolicy, ok = newdata["policy"]; !ok {
			fail(fmt.Sprintf("No 'policy' parameter supplied, expected:\n%s", expectedPolicy))
			return
		}

		if policy, ok = rawPolicy.(string); !ok {
			fail(fmt.Sprintf("policy is not a string. expected: %s, got: %v", expectedPolicy, rawPolicy))
			return
		}

		if policy != expectedPolicy {
			fail(fmt.Sprintf("Incorrect policy. Expected\n%s\n, got:\n%s", expectedPolicy, policy))
			return
		}

		success()
	}).Methods(http.MethodPut)

	router.HandleFunc(fmt.Sprintf("%s/{path}", aclBasePath), func(w http.ResponseWriter, r *http.Request) {
		var denyDelete bool

		vars := mux.Vars(r)
		path := vars["path"]

		recordPath(fmt.Sprintf("/v1/sys/policies/acl/%s", path))

		if rawDenyDelete := r.Header.Get(KVTestHeaderDenyDelete); len(rawDenyDelete) != 0 {
			var err error
			if denyDelete, err = strconv.ParseBool(rawDenyDelete); err != nil {
				fail(fmt.Sprintf("Unable to parse '%s' as boolean: %v", rawDenyDelete, err), w)
				return
			}
		}

		if denyDelete {
			fail(fmt.Sprintf("Attempt to delete policy '%s', but deletes are denied for this path", path), w)
			return
		}

		success(w)
	}).Methods(http.MethodDelete)

	return httptest.NewServer(router)
}

func TestSecretEngine_CreatePolicy(t *testing.T) {

	srv := NewFakeVaultPolicyServer(nil)
	defer srv.Close()

	tests := []struct {
		name           string
		secretEngine   *api.SecretEngine
		path           string
		expectedPolicy string
		wantErr        bool
	}{
		{
			name:           "Create policy for gcp secret engine",
			path:           "gcp",
			expectedPolicy: "gcpPolicyTest1",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcpse",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "Create policy for my-gcp-path secret engine",
			path:           "my-gcp-path",
			expectedPolicy: "gcpPolicyTest2",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcpse",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Create policy for gcp secret engine failed",
			path: "my-gcp-path",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gcpse",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{},
					},
				},
			},
			wantErr: true,
		},
		{
			name:           "KV V1 Secret Engine does not create policies",
			path:           "secret",
			expectedPolicy: "none",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kv-v1",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "KV V2 Secret Engine requires a policy for configuration",
			path:           "secret",
			expectedPolicy: "kvv2",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kv-v2",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "KV V2 Secret Engine - custom path",
			path:           "kvv2",
			expectedPolicy: "kvv2-custom-path",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kv-v2",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "kvv2",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			headers := vc.Headers()
			headers.Add(KVTestHeaderExpectedPolicy, tt.expectedPolicy)
			vc.SetHeaders(headers)

			seClient := &SecretEngine{
				secretEngine: tt.secretEngine,
				path:         tt.path,
				vaultClient:  vc,
			}
			if err := seClient.CreatePolicy(); (err != nil) != tt.wantErr {
				t.Errorf("CreatePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecretEngine_DeletePolicy(t *testing.T) {
	callRecorder := map[string]int{}

	srv := NewFakeVaultPolicyServer(&callRecorder)
	defer srv.Close()

	tests := []struct {
		name         string
		secretEngine *api.SecretEngine
		expectDelete bool
		wantErr      bool
	}{
		{
			name: "KV V1 Engine does not need to delete policies",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kv-v1",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{},
					},
				},
			},
			expectDelete: false,
			wantErr:      false,
		},
		{
			name: "KV V2 Engine deletes policy",
			secretEngine: &api.SecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kv-v2",
					Namespace: "demo",
				},
				Spec: api.SecretEngineSpec{
					VaultRef: v1.LocalObjectReference{},
					Path:     "",
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			expectDelete: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")

			headers := vc.Headers()
			headers.Add(KVTestHeaderDenyDelete, strconv.FormatBool(!tt.expectDelete))
			vc.SetHeaders(headers)

			seClient := &SecretEngine{
				secretEngine: tt.secretEngine,
				path:         "secret",
				vaultClient:  vc,
			}

			if err := seClient.DeletePolicyAndUpdateRole(); (err != nil) != tt.wantErr {
				t.Errorf("DeletePolicyAndUpdateRole() error = %v", err)
			}

			if tt.expectDelete {
				expectedPath := fmt.Sprintf("%s/k8s.-.%s.%s", aclBasePath, tt.secretEngine.ObjectMeta.Namespace, tt.secretEngine.ObjectMeta.Name)
				if callRecorder[expectedPath] == 0 {
					t.Errorf("Expected an HTTP request to '%s', but none were recorded. All recorded calls:\n%v", expectedPath, callRecorder)
				}
			}
		})
	}
}
