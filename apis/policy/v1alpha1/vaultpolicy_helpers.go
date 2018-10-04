package v1alpha1

import (
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/kubevault/operator/apis"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func (v VaultPolicy) GetKey() string {
	return v.Namespace + "/" + v.Name
}

func (v VaultPolicy) OffshootName() string {
	return v.Namespace + "-" + v.Name
}

func (v VaultPolicy) OffshootSelectors() map[string]string {
	return map[string]string{
		"app":          "vault",
		"vault_policy": v.Name,
	}
}

func (v VaultPolicy) OffshootLabels() map[string]string {
	return meta_util.FilterKeys("kubevault.com", v.OffshootSelectors(), v.Labels)
}

func (v VaultPolicy) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceVaultPolicies,
		Singular:      ResourceVaultPolicy,
		Kind:          ResourceKindVaultPolicy,
		ShortNames:    []string{"vp"},
		Categories:    []string{"vault", "policy", "appscode", "all"},
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
		SpecDefinitionName:      "github.com/kubevault/operator/apis/policy/v1alpha1.VaultPolicy",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: apis.EnableStatusSubresource,
		AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
			{
				Name:     "Status",
				Type:     "string",
				JSONPath: ".status.status",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
		},
	}, apis.SetNameSchema)
}

func (v VaultPolicy) IsValid() error {
	return nil
}
