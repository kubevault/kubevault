package database

import (
	"encoding/json"
	"fmt"

	patchutilv1 "github.com/appscode/kutil/core/v1"
	patchutil "github.com/appscode/kutil/rbac/v1"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	crd "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/kubevault/operator/pkg/vault"
	vaultcs "github.com/kubevault/operator/pkg/vault"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

type DBCredManager struct {
	dbAccessReq *api.DatabaseAccessRequest
	kubeClient  kubernetes.Interface
	vaultClient *vaultapi.Client
	path        string
	roleName    string
}

func NewDatabaseCredentialManager(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, dbAR *api.DatabaseAccessRequest) (DatabaseCredentialManager, error) {
	v, dbPath, roleName, err := GetVaultClientDBPathAndRole(kClient, appClient, cr, dbAR.Spec.RoleRef)
	if err != nil {
		return nil, err
	}
	return &DBCredManager{
		dbAccessReq: dbAR,
		roleName:    roleName,
		kubeClient:  kClient,
		vaultClient: v,
		path:        dbPath,
	}, nil
}

func GetVaultClientDBPathAndRole(kClient kubernetes.Interface, appClient appcat_cs.AppcatalogV1alpha1Interface, cr crd.Interface, ref api.RoleReference) (*vaultapi.Client, string, string, error) {
	var (
		vaultRef *appcat.AppReference
		roleName string
	)

	switch ref.Kind {
	case api.ResourceKindMongoDBRole:
		r, err := cr.AuthorizationV1alpha1().MongoDBRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "MongoDBRole %s/%s", ref.Namespace, ref.Name)
		}
		vaultRef = r.Spec.AuthManagerRef
		roleName = r.RoleName()

	case api.ResourceKindMySQLRole:
		r, err := cr.AuthorizationV1alpha1().MySQLRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "MySQLRole %s/%s", ref.Namespace, ref.Name)
		}
		vaultRef = r.Spec.AuthManagerRef
		roleName = r.RoleName()

	case api.ResourceKindPostgresRole:
		r, err := cr.AuthorizationV1alpha1().PostgresRoles(ref.Namespace).Get(ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, "", "", errors.Wrapf(err, "PostgresRole %s/%s", ref.Namespace, ref.Name)
		}
		vaultRef = r.Spec.AuthManagerRef
		roleName = r.RoleName()

	default:
		return nil, "", "", errors.Errorf("unknown or unsupported role kind '%s'", ref.Kind)
	}

	v, err := vaultcs.NewClient(kClient, appClient, vaultRef)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "failed to create vault client")
	}
	path, err := getDatabasePath(appClient, *vaultRef)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "failed to get database path")
	}
	return v, path, roleName, nil
}

// Creates a kubernetes secret containing database credential
func (d *DBCredManager) CreateSecret(name string, namespace string, cred *vault.DatabaseCredential) error {
	data := map[string][]byte{}
	if cred != nil {
		data = map[string][]byte{
			"username": []byte(cred.Data.Username),
			"password": []byte(cred.Data.Password),
		}
	}

	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	_, _, err := patchutilv1.CreateOrPatchSecret(d.kubeClient, obj, func(s *corev1.Secret) *corev1.Secret {
		s.Data = data
		addOwnerRefToObject(s, d.AsOwner())
		return s
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create/update secret %s/%s", namespace, name)
	}
	return nil
}

// Creates kubernetes role
func (d *DBCredManager) CreateRole(name string, namespace string, secretName string) error {
	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	_, _, err := patchutil.CreateOrPatchRole(d.kubeClient, obj, func(role *rbacv1.Role) *rbacv1.Role {
		role.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"", // represents core api
				},
				Resources: []string{
					"secrets",
				},
				ResourceNames: []string{
					secretName,
				},
				Verbs: []string{
					"get",
				},
			},
		}

		addOwnerRefToObject(role, d.AsOwner())
		return role
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create rbac role %s/%s", namespace, name)
	}
	return nil
}

// Create kubernetes role binding
func (d *DBCredManager) CreateRoleBinding(name string, namespace string, roleName string, subjects []rbacv1.Subject) error {
	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	_, _, err := patchutil.CreateOrPatchRoleBinding(d.kubeClient, obj, func(role *rbacv1.RoleBinding) *rbacv1.RoleBinding {
		role.RoleRef = rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleName,
		}
		role.Subjects = subjects

		addOwnerRefToObject(role, d.AsOwner())
		return role
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create/update rbac role binding %s/%s", namespace, name)
	}
	return nil
}

// https://www.vaultproject.io/api/system/leases.html#read-lease
//
// Whether or not lease is expired in vault
// In vault, lease is revoked if lease is expired
func (d *DBCredManager) IsLeaseExpired(leaseID string) (bool, error) {
	if leaseID == "" {
		return true, nil
	}

	req := d.vaultClient.NewRequest("PUT", "/v1/sys/leases/lookup")
	err := req.SetJSONBody(map[string]string{
		"lease_id": leaseID,
	})
	if err != nil {
		return false, errors.WithStack(err)
	}

	resp, err := d.vaultClient.RawRequest(req)
	if resp == nil && err != nil {
		return false, errors.WithStack(err)
	}

	defer resp.Body.Close()
	errResp := vaultapi.ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if len(errResp.Errors) > 0 {
		return true, nil
	}
	return false, nil
}

// RevokeLease revokes respective lease
// It's safe to call multiple time. It doesn't give
// error even if respective lease_id doesn't exist
// but it will give an error if lease_id is empty
func (d *DBCredManager) RevokeLease(leaseID string) error {
	err := d.vaultClient.Sys().Revoke(leaseID)
	if err != nil {
		return errors.Wrap(err, "failed to revoke lease")
	}
	return nil
}

// Gets credential from vault
func (p *DBCredManager) GetCredential() (*vault.DatabaseCredential, error) {
	path := fmt.Sprintf("/v1/%s/creds/%s", p.path, p.roleName)
	req := p.vaultClient.NewRequest("GET", path)

	resp, err := p.vaultClient.RawRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get mysql credential")
	}

	cred := vault.DatabaseCredential{}

	err = json.NewDecoder(resp.Body).Decode(&cred)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode json from mysql credential response")
	}
	return &cred, nil
}

// asOwner returns an owner reference
func (p *DBCredManager) AsOwner() metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.ResourceKindDatabaseAccessRequest,
		Name:       p.dbAccessReq.Name,
		UID:        p.dbAccessReq.UID,
		Controller: &trueVar,
	}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	if !IsOwnerRefAlreadyExists(o, r) {
		o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
	}
}

func IsOwnerRefAlreadyExists(o metav1.Object, r metav1.OwnerReference) bool {
	refs := o.GetOwnerReferences()
	for _, u := range refs {
		if u.Name != r.Name &&
			u.UID == r.UID &&
			u.Kind == r.Kind &&
			u.APIVersion == u.APIVersion {
			return true
		}
	}
	return false
}
