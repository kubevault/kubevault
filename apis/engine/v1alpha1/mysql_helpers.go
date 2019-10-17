package v1alpha1

import (
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdutils "kmodules.xyz/client-go/apiextensions/v1beta1"
)

const DefaultMySQLDatabasePlugin = "mysql-database-plugin"

func (r MySQLRole) RoleName() string {
	cluster := "-"
	if r.ClusterName != "" {
		cluster = r.ClusterName
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, r.Namespace, r.Name)
}

func (r MySQLRole) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourceMySQLRoles,
		Singular:      ResourceMySQLRole,
		Kind:          ResourceKindMySQLRole,
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
		SpecDefinitionName:      "kubevault.dev/operator/apis/engine/v1alpha1.MySQLRole",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: true,
	})
}

func (r MySQLRole) IsValid() error {
	return nil
}

func (m *MySQLConfiguration) SetDefaults() {
	if m == nil {
		return
	}

	// If user doesn't specify the list of AllowedRoles
	// It is set to "*" (allow all)
	if m.AllowedRoles == nil || len(m.AllowedRoles) == 0 {
		m.AllowedRoles = []string{"*"}
	}

	if m.PluginName == "" {
		m.PluginName = DefaultMySQLDatabasePlugin
	}
}
