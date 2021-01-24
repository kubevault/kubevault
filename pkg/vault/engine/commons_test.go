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
	"net/http"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

const KVTestHeaderExpectedVersion = "X-Vault-KV-Expected-Version"
const KVTestHeaderExpectedMaxVersions = "X-Vault-KV-Expected-Max-Version"
const KVTestHeaderExpectedCasRequired = "X-Vault-KV-Expected-Cas-Required"
const KVTestHeaderExpectedDeleteVersionsAfter = "X-Vault-KV-Expected-Delete-Versions-After"
const KVTestHeaderExpectedPolicy = "X-Vault-Expected-Policy-Name"
const KVTestHeaderDenyDelete = "X-Vault-Deny-Delete"

func mustWrite(b []byte, w http.ResponseWriter) {
	_, err := w.Write(b)
	utilruntime.Must(err)
}

func mustWriteString(s string, w http.ResponseWriter) {
	mustWrite([]byte(s), w)
}

func fail(message string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	mustWriteString(message, w)
	mustWriteString("\n", w)
}

func success(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
