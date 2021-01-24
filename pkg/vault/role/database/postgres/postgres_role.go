/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package postgres

import (
	"context"
	"fmt"

	api "kubevault.dev/apimachinery/apis/engine/v1alpha1"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type PostgresRole struct {
	pgRole       *api.PostgresRole
	vaultClient  *vaultapi.Client
	kubeClient   kubernetes.Interface
	dbBinding    *appcat.AppBinding
	databasePath string
}

func NewPostgresRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, v *vaultapi.Client, pgRole *api.PostgresRole, databasePath string) (*PostgresRole, error) {
	ref := pgRole.Spec.DatabaseRef
	dbBinding, err := appClient.AppBindings(ref.Namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &PostgresRole{
		pgRole:       pgRole,
		vaultClient:  v,
		kubeClient:   kClient,
		dbBinding:    dbBinding,
		databasePath: databasePath,
	}, nil
}

// https://www.vaultproject.io/api/secret/databases/index.html#create-role
//
// CreateRole creates role
func (p *PostgresRole) CreateRole() error {
	name := p.pgRole.RoleName()
	pg := p.pgRole.Spec

	path := fmt.Sprintf("/v1/%s/roles/%s", p.databasePath, name)
	req := p.vaultClient.NewRequest("POST", path)

	var dbName string
	if pg.DatabaseRef != nil {
		if pg.DatabaseRef.Name == "" {
			return errors.New("DatabaseRef.Name is empty")
		}
		if pg.DatabaseRef.Namespace == "" {
			return errors.New("DatabaseRef.Namespace is empty")
		}
		dbName = api.GetDBNameFromAppBindingRef(pg.DatabaseRef)
	} else if pg.DatabaseName != "" {
		dbName = pg.DatabaseName
	} else {
		return errors.New("both DatabaseRef and DatabaseName are empty")
	}
	payload := map[string]interface{}{
		"db_name":             dbName,
		"creation_statements": pg.CreationStatements,
	}

	if len(pg.RevocationStatements) > 0 {
		payload["revocation_statements"] = pg.RevocationStatements
	}
	if len(pg.RollbackStatements) > 0 {
		payload["rollback_statements"] = pg.RollbackStatements
	}
	if len(pg.RenewStatements) > 0 {
		payload["renew_statements"] = pg.RenewStatements
	}
	if pg.DefaultTTL != "" {
		payload["default_ttl"] = pg.DefaultTTL
	}
	if pg.MaxTTL != "" {
		payload["max_ttl"] = pg.MaxTTL
	}

	err := req.SetJSONBody(payload)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = p.vaultClient.RawRequest(req)
	if err != nil {
		return errors.Wrapf(err, "failed to create database role %s for config %s", name, pg.DatabaseRef.Name)
	}
	return nil
}
