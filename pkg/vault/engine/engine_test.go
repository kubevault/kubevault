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

package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	api "kubevault.dev/operator/apis/engine/v1alpha1"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const sampleMountResponse = `
{
  "sys/": {
    "accessor": "system_fa208b63",
    "config": {
      "default_lease_ttl": 0,
      "force_no_cache": false,
      "max_lease_ttl": 0,
      "passthrough_request_headers": [
        "Accept"
      ]
    },
    "description": "system endpoints used for control, policy and debugging",
    "local": false,
    "options": null,
    "seal_wrap": false,
    "type": "system",
    "uuid": "65a2be4c-2381-d55d-3ae6-f84782829dfc"
  },
  "identity/": {
    "accessor": "identity_3b96d858",
    "config": {
      "default_lease_ttl": 0,
      "force_no_cache": false,
      "max_lease_ttl": 0
    },
    "description": "identity store",
    "local": false,
    "options": null,
    "seal_wrap": false,
    "type": "identity",
    "uuid": "92da81eb-997c-4445-dd41-4689a486f160"
  },
  "cubbyhole/": {
    "accessor": "cubbyhole_0bacc8f3",
    "config": {
      "default_lease_ttl": 0,
      "force_no_cache": false,
      "max_lease_ttl": 0
    },
    "description": "per-token private secret storage",
    "local": true,
    "options": null,
    "seal_wrap": false,
    "type": "cubbyhole",
    "uuid": "9ba524f2-645b-274d-b64a-89f78ba36fc0"
  },
  "request_id": "e5d01697-13c5-f9b6-cf55-b4c00807c162",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "cubbyhole/": {
      "accessor": "cubbyhole_0bacc8f3",
      "config": {
        "default_lease_ttl": 0,
        "force_no_cache": false,
        "max_lease_ttl": 0
      },
      "description": "per-token private secret storage",
      "local": true,
      "options": null,
      "seal_wrap": false,
      "type": "cubbyhole",
      "uuid": "9ba524f2-645b-274d-b64a-89f78ba36fc0"
    },
    "identity/": {
      "accessor": "identity_3b96d858",
      "config": {
        "default_lease_ttl": 0,
        "force_no_cache": false,
        "max_lease_ttl": 0
      },
      "description": "identity store",
      "local": false,
      "options": null,
      "seal_wrap": false,
      "type": "identity",
      "uuid": "92da81eb-997c-4445-dd41-4689a486f160"
    },
    "sys/": {
      "accessor": "system_fa208b63",
      "config": {
        "default_lease_ttl": 0,
        "force_no_cache": false,
        "max_lease_ttl": 0,
        "passthrough_request_headers": [
          "Accept"
        ]
      },
      "description": "system endpoints used for control, policy and debugging",
      "local": false,
      "options": null,
      "seal_wrap": false,
      "type": "system",
      "uuid": "65a2be4c-2381-d55d-3ae6-f84782829dfc"
    }
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
`

func NewFakeVaultMountServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/v1/sys/mounts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(sampleMountResponse))
		utilruntime.Must(err)
	}).Methods(http.MethodGet)

	router.HandleFunc("/v1/sys/mounts/{path}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := vars["path"]
		data, _ := ioutil.ReadAll(r.Body)

		fail := func(message string) {
			fail(message, w)
			mustWrite(data, w)
		}

		var newdata map[string]interface{}
		_ = json.Unmarshal(data, &newdata)
		if engineType, ok := newdata["type"]; ok {
			fmt.Printf("'%s': '%s'\n", engineType, path)

			if engineType == api.EngineTypeKV {
				if options := newdata["options"]; options != nil {
					if optionMap, ok := options.(map[string]interface{}); ok {
						if version, ok := optionMap["version"]; ok {
							expectedVersion := r.Header.Get(KVTestHeaderExpectedVersion)
							if expectedVersion == "" {
								fail("unable to verify version; expected version header is missing")
								return
							}
							fmt.Println("  expected version:", expectedVersion)
							if expectedVersion != version {
								fail(fmt.Sprintf("expected version %v, got: %v", expectedVersion, version))
								return
							}
						} else {
							// Technically, it's not, but enabling the KV engine at v1 when the
							// desired vrsion is v2 could lead to a situation where clients
							// expecting a v2 store would be very disappointed, since the API is
							// different between the two. Also, upgrading from a v1 to a v2 makes
							// the mount unavailable during the upgrade, so it's best to just always
							// specify the version at creation time.
							fail("version is required for KV engine")
							return
						}
					} else {
						fail(fmt.Sprintf("options should be a map: %v", options))
					}
				} else {
					// Because it's got the version in it. See above for why version is required.
					fail("options is required for KV engine")
					return
				}

				if path == DefaultKVPath {
					w.WriteHeader(http.StatusOK)
				}

			} else if engineType == path {
				w.WriteHeader(http.StatusOK)
			} else {
				fail("path did not match engineType")
			}
		}
	}).Methods(http.MethodPost)

	return httptest.NewServer(router)
}

func TestSecretEngine_EnableSecretEngine(t *testing.T) {
	srv := NewFakeVaultMountServer()
	defer srv.Close()

	tests := []struct {
		name         string
		secretEngine *api.SecretEngine
		wantErr      bool
		path         string
		extraHeaders map[string]string
	}{
		{
			name: "enable gcp secret engine: successful",
			path: "gcp",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "enable aws secret engine: successful",
			path: "aws",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "enable azure secret engine: successful",
			path: "azure",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Azure: &api.AzureConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "enable database secret engine: successful",
			path: "database",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						Postgres: &api.PostgresConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "enable database secret engine: successful",
			path: "database",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						MySQL: &api.MySQLConfiguration{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "enable database secret engine: unsuccessful",
			path: "database",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						GCP: &api.GCPConfiguration{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "enable azure secret engine: unsuccessful",
			path: "azure",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "enable kv-v1 secret engine: successful",
			path: "secret",
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion: strconv.FormatInt(DefaultKVVersion, 10),
			},
		},
		{
			name: "enable kv-v1 secret engine: unsuccessful",
			path: DefaultKVPath,
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						AWS: &api.AWSConfiguration{},
					},
				},
			},
			wantErr: true,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion: "1",
			},
		},
		{
			name: "enable kv-v2 secret engine: successful",
			path: DefaultKVPath,
			secretEngine: &api.SecretEngine{
				Spec: api.SecretEngineSpec{
					SecretEngineConfiguration: api.SecretEngineConfiguration{
						KV: &api.KVConfiguration{
							Version: 2,
						},
					},
				},
			},
			wantErr: false,
			extraHeaders: map[string]string{
				KVTestHeaderExpectedVersion: "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			vc, err := vaultClient(srv.URL)
			assert.Nil(t, err, "failed to create vault client")
			if tt.extraHeaders != nil {
				// As of Vault 1.4, we could just use vc.AddHeader
				vcHeaders := vc.Headers()
				for key, value := range tt.extraHeaders {
					vcHeaders.Add(key, value)
				}
				vc.SetHeaders(vcHeaders)
			}

			seClient := &SecretEngine{
				secretEngine: tt.secretEngine,
				vaultClient:  vc,
				path:         tt.path,
			}
			if err := seClient.EnableSecretEngine(); (err != nil) != tt.wantErr {
				t.Errorf("EnableSecretEngine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
