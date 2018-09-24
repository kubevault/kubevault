package v1alpha1

import (
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	meta_util "github.com/appscode/kutil/meta"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func (v VaultServer) GetKey() string {
	return v.Namespace + "/" + v.Name
}

func (v VaultServer) OffshootName() string {
	return v.Name
}

func (v VaultServer) OffshootSelectors() map[string]string {
	return map[string]string{
		"app":           "vault",
		"vault_cluster": v.Name,
	}
}

func (v VaultServer) OffshootLabels() map[string]string {
	return meta_util.FilterKeys("kubevault.com", v.OffshootSelectors(), v.Labels)
}

func (v VaultServer) ConfigMapName() string {
	return v.OffshootName() + "-vault-config"
}

func (v VaultServer) TLSSecretName() string {
	return v.OffshootName() + "-vault-tls"
}

func (v VaultServer) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceVaultServers,
		Singular:      ResourceVaultServer,
		Kind:          ResourceKindVaultServer,
		ShortNames:    []string{"vs"},
		Categories:    []string{"vault", "appscode", "all"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "vault"},
		},
		SpecDefinitionName:      "github.com/kubevault/operator/apis/kubevault/v1alpha1.VaultServer",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
		AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
			{
				Name:     "Nodes",
				Type:     "string",
				JSONPath: ".spec.nodes",
			},
			{
				Name:     "Version",
				Type:     "string",
				JSONPath: ".spec.version",
			},
			{
				Name:     "Status",
				Type:     "string",
				JSONPath: ".status.phase",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
		},
	}, setNameSchema)
}

func (v VaultServer) IsValid() error {
	return nil
}
