package v1alpha1

import (
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/clusterid"
	"kubevault.dev/operator/apis"
)

func (v VaultPolicyBinding) GetKey() string {
	return ResourceVaultPolicyBinding + "/" + v.Namespace + "/" + v.Name
}

func (v VaultPolicyBinding) PolicyBindingName() string {
	if v.Spec.RoleName != "" {
		return v.Spec.RoleName
	}

	cluster := "-"
	if clusterid.ClusterName() != "" {
		cluster = clusterid.ClusterName()
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, v.Namespace, v.Name)
}

func (v VaultPolicyBinding) OffshootSelectors() map[string]string {
	return map[string]string{
		"app":                  "vault",
		"vault_policy_binding": v.Name,
	}
}

func (v VaultPolicyBinding) OffshootLabels() map[string]string {
	return meta_util.FilterKeys("kubevault.com", v.OffshootSelectors(), v.Labels)
}

func (v VaultPolicyBinding) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceVaultPolicyBindings,
		Singular:      ResourceVaultPolicyBinding,
		Kind:          ResourceKindVaultPolicyBinding,
		ShortNames:    []string{"vpb"},
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
		SpecDefinitionName:      "kubevault.dev/operator/apis/policy/v1alpha1.VaultPolicyBinding",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: apis.EnableStatusSubresource,
		AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
			{
				Name:     "Phase",
				Type:     "string",
				JSONPath: ".status.phase",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
		},
	}, apis.SetNameSchema)
}

func (v VaultPolicyBinding) IsValid() error {
	return nil
}

func (v *VaultPolicyBinding) SetDefaults() {
	if v == nil {
		return
	}

	if v.Spec.RoleName == "" {
		v.Spec.RoleName = v.PolicyBindingName()
	}

	if v.Spec.AuthPath == "" {
		v.Spec.AuthPath = "kubernetes"
	}
}
