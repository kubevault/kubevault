package v1alpha1

import (
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
	"kubevault.dev/operator/apis"
)

func (d AWSAccessKeyRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceAWSAccessKeyRequests,
		Singular:      ResourceAWSAccessKeyRequest,
		Kind:          ResourceKindAWSAccessKeyRequest,
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
		SpecDefinitionName:      "kubevault.dev/operator/apis/kubevault/v1alpha1.AWSAccessKeyRequest",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: apis.EnableStatusSubresource,
	})
}

func (d AWSAccessKeyRequest) IsValid() error {
	return nil
}
