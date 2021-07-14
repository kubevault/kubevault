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
	"errors"
	"fmt"
	"path/filepath"

	"kubevault.dev/apimachinery/apis"
	"kubevault.dev/apimachinery/apis/kubevault"
	"kubevault.dev/apimachinery/crds"

	"k8s.io/apimachinery/pkg/labels"
	appslister "k8s.io/client-go/listers/apps/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	apps_util "kmodules.xyz/client-go/apps/v1"
	"kmodules.xyz/client-go/meta"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
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

func (v VaultServer) ConfigSecretName() string {
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

func (v VaultServer) ServiceName(alias ServiceAlias) string {
	if alias == VaultServerServiceVault {
		return v.Name
	}
	return meta_util.NameWithSuffix(v.Name, string(alias))
}

func (v VaultServer) StatsLabels() map[string]string {
	labels := v.OffshootLabels()
	labels["feature"] = "stats"
	return labels
}

// Returns the default certificate secret name for given alias.
func (vs *VaultServer) DefaultCertSecretName(alias string) string {
	return meta.NameWithSuffix(fmt.Sprintf("%s-%s", vs.Name, alias), "certs")
}

// Returns certificate secret name for given alias if exists,
// otherwise returns the default certificate secret name.
func (vs *VaultServer) GetCertSecretName(alias string) string {
	if vs.Spec.TLS != nil {
		sName, valid := kmapi.GetCertificateSecretName(vs.Spec.TLS.Certificates, alias)
		if valid {
			return sName
		}
	}

	return vs.DefaultCertSecretName(alias)
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

func (vs *VaultServer) GetCertificateCN(alias VaultCertificateAlias) string {
	return fmt.Sprintf("%s-%s", vs.Name, string(alias))
}

func (vs *VaultServer) Scheme() string {
	if vs.Spec.TLS != nil {
		return "https"
	}
	return "http"
}

func (vsb *BackendStorageSpec) GetBackendType() (VaultServerBackend, error) {
	switch {
	case vsb.Inmem != nil:
		return VaultServerInmem, nil
	case vsb.Etcd != nil:
		return VaultServerEtcd, nil
	case vsb.Gcs != nil:
		return VaultServerGcs, nil
	case vsb.S3 != nil:
		return VaultServerS3, nil
	case vsb.Azure != nil:
		return VaultServerAzure, nil
	case vsb.PostgreSQL != nil:
		return VaultServerPostgreSQL, nil
	case vsb.MySQL != nil:
		return VaultServerMySQL, nil
	case vsb.File != nil:
		return VaultServerFile, nil
	case vsb.DynamoDB != nil:
		return VaultServerDynamoDB, nil
	case vsb.Swift != nil:
		return VaultServerSwift, nil
	case vsb.Consul != nil:
		return VaultServerConsul, nil
	case vsb.Raft != nil:
		return VaultServerRaft, nil
	default:
		return "", errors.New("unknown backened type")
	}
}

func (v *VaultServer) CertificateMountPath(alias VaultCertificateAlias) string {
	return filepath.Join(apis.CertificatePath, string(alias))
}

func (v *VaultServer) ReplicasAreReady(lister appslister.StatefulSetLister) (bool, string, error) {
	// Desired number of statefulSets
	expectedItems := 1
	return checkReplicas(lister.StatefulSets(v.Namespace), labels.SelectorFromSet(v.OffshootLabels()), expectedItems)
}

func checkReplicas(lister appslister.StatefulSetNamespaceLister, selector labels.Selector, expectedItems int) (bool, string, error) {
	items, err := lister.List(selector)
	if err != nil {
		return false, "", err
	}

	if len(items) < expectedItems {
		return false, fmt.Sprintf("All StatefulSets are not available. Desire number of StatefulSet: %d, Available: %d", expectedItems, len(items)), nil
	}

	// return isReplicasReady, message, error
	ready, msg := apps_util.StatefulSetsAreReady(items)
	return ready, msg, nil
}

// GetServiceTemplate returns a pointer to the desired serviceTemplate referred by "alias". Otherwise, it returns nil.
func (vs *VaultServer) GetServiceTemplate(alias ServiceAlias) ofst.ServiceTemplateSpec {
	templates := vs.Spec.ServiceTemplates
	for i := range templates {
		c := templates[i]
		if c.Alias == alias {
			return c.ServiceTemplateSpec
		}
	}
	return ofst.ServiceTemplateSpec{}
}
