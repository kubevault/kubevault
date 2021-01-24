/*
Copyright AppsCode Inc. and Contributors

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

package v1alpha1

import (
	"fmt"

	"kubevault.dev/apimachinery/apis/kubevault"
	"kubevault.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (_ VaultServer) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceVaultServers))
}

func (_ VaultServer) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourceVaultServers, kubevault.GroupName)
}

func (v VaultServer) GetKey() string {
	return v.Namespace + "/" + v.Name
}

func (v VaultServer) OffshootName() string {
	return v.Name
}

func (v VaultServer) ServiceAccountName() string {
	return v.Name
}

func (v VaultServer) ServiceAccountForTokenReviewer() string {
	return meta_util.NameWithSuffix(v.Name, "k8s-token-reviewer")
}

func (v VaultServer) PolicyNameForPolicyController() string {
	return meta_util.NameWithSuffix(v.Name, "policy-controller")
}

func (v VaultServer) PolicyNameForAuthMethodController() string {
	return meta_util.NameWithSuffix(v.Name, "auth-method-controller")
}

func (v VaultServer) AppBindingName() string {
	return v.Name
}

func (v VaultServer) OffshootSelectors() map[string]string {
	return map[string]string{
		meta_util.NameLabelKey:      v.ResourceFQN(),
		meta_util.InstanceLabelKey:  v.Name,
		meta_util.ManagedByLabelKey: kubevault.GroupName,
	}
}

func (v VaultServer) OffshootLabels() map[string]string {
	return meta_util.FilterKeys("kubevault.com", v.OffshootSelectors(), v.Labels)
}

func (v VaultServer) ConfigMapName() string {
	return meta_util.NameWithSuffix(v.Name, "vault-config")
}

func (v VaultServer) TLSSecretName() string {
	return meta_util.NameWithSuffix(v.Name, "vault-tls")
}

func (v VaultServer) IsValid() error {
	return nil
}

func (v VaultServer) StatsServiceName() string {
	return meta_util.NameWithSuffix(v.Name, "stats")
}

func (v VaultServer) StatsLabels() map[string]string {
	labels := v.OffshootLabels()
	labels["feature"] = "stats"
	return labels
}

func (v VaultServer) StatsService() mona.StatsAccessor {
	return &vaultServerStatsService{&v}
}

type vaultServerStatsService struct {
	*VaultServer
}

func (e vaultServerStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return e.VaultServer.OffshootLabels()
}

func (e vaultServerStatsService) GetNamespace() string {
	return e.VaultServer.GetNamespace()
}

func (e vaultServerStatsService) ServiceName() string {
	return e.StatsServiceName()
}

func (e vaultServerStatsService) ServiceMonitorName() string {
	return e.ServiceName()
}

func (e vaultServerStatsService) Path() string {
	return "/metrics"
}

func (e vaultServerStatsService) Scheme() string {
	return ""
}
