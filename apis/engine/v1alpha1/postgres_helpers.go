package v1alpha1

import (
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
	"kubevault.dev/operator/apis"
)

const DefaultPostgresDatabasePlugin = "postgresql-database-plugin"

func (r PostgresRole) RoleName() string {
	cluster := "-"
	if r.ClusterName != "" {
		cluster = r.ClusterName
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, r.Namespace, r.Name)
}

func (r PostgresRole) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePostgresRoles,
		Singular:      ResourcePostgresRole,
		Kind:          ResourceKindPostgresRole,
		Categories:    []string{"datastore", "kubedb", "appscode", "all"},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubedb"},
		},
		SpecDefinitionName:      "kubevault.dev/operator/apis/engine/v1alpha1.PostgresRole",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: apis.EnableStatusSubresource,
	})
}

func (r PostgresRole) IsValid() error {
	return nil
}

func (p *PostgresConfiguration) SetDefaults() {
	if p == nil {
		return
	}

	// If user doesn't specify the list of AllowedRoles
	// It is set to "*" (allow all)
	if p.AllowedRoles == nil || len(p.AllowedRoles) == 0 {
		p.AllowedRoles = []string{"*"}
	}

	if p.PluginName == "" {
		p.PluginName = DefaultPostgresDatabasePlugin
	}
}
