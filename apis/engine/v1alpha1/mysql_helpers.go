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

package v1alpha1

import (
	"fmt"

	"kubevault.dev/operator/api/crds"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func (_ MySQLRole) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceMySQLRoles))
}

const DefaultMySQLDatabasePlugin = "mysql-database-plugin"

func (r MySQLRole) RoleName() string {
	cluster := "-"
	if r.ClusterName != "" {
		cluster = r.ClusterName
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, r.Namespace, r.Name)
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
