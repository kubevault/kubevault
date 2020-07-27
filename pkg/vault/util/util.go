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

package util

import (
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
)

// VaultServiceURL returns the DNS record of the vault service in the given namespace.
func VaultServiceURL(name, namespace string, port int) string {
	return fmt.Sprintf("https://%s.%s.svc:%d", name, namespace, port)
}

// PodDNSName constructs the dns name on which a pod can be addressed
func PodDNSName(p core.Pod) string {
	podIP := strings.Replace(p.Status.PodIP, ".", "-", -1)
	return fmt.Sprintf("%s.%s.pod", podIP, p.Namespace)
}
