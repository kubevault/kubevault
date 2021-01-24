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

	"kubevault.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (_ MongoDBRole) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceMongoDBRoles))
}

const DefaultMongoDBDatabasePlugin = "mongodb-database-plugin"

func (r MongoDBRole) RoleName() string {
	cluster := "-"
	if r.ClusterName != "" {
		cluster = r.ClusterName
	}
	return fmt.Sprintf("k8s.%s.%s.%s", cluster, r.Namespace, r.Name)
}

func (r MongoDBRole) IsValid() error {
	return nil
}

func (m *MongoDBConfiguration) SetDefaults() {
	if m == nil {
		return
	}

	// If user doesn't specify the list of AllowedRoles
	// It is set to "*" (allow all)
	if m.AllowedRoles == nil || len(m.AllowedRoles) == 0 {
		m.AllowedRoles = []string{"*"}
	}

	if m.PluginName == "" {
		m.PluginName = DefaultMongoDBDatabasePlugin
	}
}
