package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	"github.com/kube-vault/operator/pkg/vault/storage"
	"github.com/kube-vault/operator/pkg/vault/util"
	"github.com/stretchr/testify/assert"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

const (
	vaultTestName      = "ex-vault-unit-test"
	vaultTestNamespace = "ns-vault-test"
	vaultTestNodes     = 3
	vaultTestUID       = "vault-test-1204"
)

func getConfigData(t *testing.T, extraConfig string, strg *api.BackendStorageSpec) string {
	cfg := util.GetListenerConfig()
	if len(extraConfig) != 0 {
		cfg = fmt.Sprintf("%s\n%s", cfg, extraConfig)
	}

	strgSrv, err := storage.NewStorage(strg)
	assert.Nil(t, err)

	storageCfg, err := strgSrv.GetStorageConfig()
	if err != nil {
		t.Fatal("create vault storage config failed", err)
	}
	cfg = fmt.Sprintf("%s\n%s", cfg, storageCfg)

	return cfg
}

func getVaultObjectMeta(i int) metav1.ObjectMeta {
	suffix := strconv.Itoa(i)
	return metav1.ObjectMeta{
		Name:      vaultTestName + suffix,
		Namespace: vaultTestNamespace + suffix,
	}
}

func getConfigMap(meta metav1.ObjectMeta, data string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name + "-config",
			Namespace: meta.Namespace,
		},
		Data: map[string]string{
			"vault.hcl": data,
		},
	}
}

func createConfigMap(t *testing.T, client kubernetes.Interface, cm *corev1.ConfigMap) {
	_, err := client.CoreV1().ConfigMaps(cm.Namespace).Create(cm)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteConfigMap(t *testing.T, client kubernetes.Interface, cm *corev1.ConfigMap) {
	err := client.CoreV1().ConfigMaps(cm.Namespace).Delete(cm.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func getSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VaultTlsSecretName,
			Namespace: vaultTestNamespace,
		},
	}
}

func createSecret(t *testing.T, client kubernetes.Interface, s *corev1.Secret) {
	_, err := client.CoreV1().Secrets(s.Namespace).Create(s)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteSecret(t *testing.T, client kubernetes.Interface, s *corev1.Secret) {
	err := client.CoreV1().Secrets(s.Namespace).Delete(s.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func getDeployment(t *testing.T, meta metav1.ObjectMeta, image string, replica ...int32) *appsv1beta1.Deployment {
	var r int32
	if len(replica) == 0 {
		r = 0
	} else {
		r = replica[0]
	}
	return &appsv1beta1.Deployment{
		ObjectMeta: meta,
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &r,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: image,
					}},
				},
			},
			Strategy: appsv1beta1.DeploymentStrategy{
				Type: appsv1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1beta1.RollingUpdateDeployment{
					MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				},
			},
		},
	}
}

func createDeployment(t *testing.T, client kubernetes.Interface, d *appsv1beta1.Deployment) {
	_, err := client.AppsV1beta1().Deployments(d.Namespace).Create(d)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteDeployment(t *testing.T, client kubernetes.Interface, d *appsv1beta1.Deployment) {
	err := client.AppsV1beta1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func createService(t *testing.T, client kubernetes.Interface, s *corev1.Service) {
	_, err := client.CoreV1().Services(s.Namespace).Create(s)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteService(t *testing.T, client kubernetes.Interface, s *corev1.Service) {
	err := client.CoreV1().Services(s.Namespace).Delete(s.Name, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrepareConfig(t *testing.T) {
	var (
		backendInmem = api.BackendStorageSpec{
			Inmem: true,
		}

		backendEtcd = api.BackendStorageSpec{
			Etcd: &api.EtcdSpec{
				Address:      "localhost:2379",
				EtcdApi:      "v3",
				DiscoverySrv: "e.com",
				Path:         "/path",
			},
		}
		backendEtcdWithSecret = api.BackendStorageSpec{
			Etcd: &api.EtcdSpec{
				Address:       "localhost:2379",
				EtcdApi:       "v3",
				DiscoverySrv:  "e.com",
				Path:          "/path",
				TLSSecretName: VaultTlsSecretName,
			},
		}

		extraConfig = `
telemetry {
  statsite_address = "statsite.test.local:8125"
}
`
	)

	testData := []struct {
		name            string
		vs              api.VaultServer
		extraConfigM    *corev1.ConfigMap
		extraSecret     *corev1.Secret
		exptErr         bool
		exptConfigMData map[string]string
	}{
		{
			"backend inmem, no extra config",
			api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
				Spec: api.VaultServerSpec{
					BackendStorage: backendInmem,
				},
			},
			nil,
			nil,
			false,
			map[string]string{filepath.Base(util.VaultConfigFile): getConfigData(t, "", &backendInmem)},
		},
		{
			"backend inmem, with extra config",
			api.VaultServer{
				ObjectMeta: getVaultObjectMeta(2),
				Spec: api.VaultServerSpec{
					BackendStorage: backendInmem,
					ConfigMapName:  getVaultObjectMeta(2).Name + "-config",
				},
			},
			getConfigMap(getVaultObjectMeta(2), extraConfig),
			nil,
			false,
			map[string]string{filepath.Base(util.VaultConfigFile): getConfigData(t, extraConfig, &backendInmem)},
		},
		{
			"backend etcd without tls secret, no extra config",
			api.VaultServer{
				ObjectMeta: getVaultObjectMeta(3),
				Spec: api.VaultServerSpec{
					BackendStorage: backendEtcd,
				},
			},
			nil,
			nil,
			false,
			map[string]string{filepath.Base(util.VaultConfigFile): getConfigData(t, "", &backendEtcd)},
		},
		{
			"backend etcd with tls secret, no extra config",
			api.VaultServer{
				ObjectMeta: getVaultObjectMeta(4),
				Spec: api.VaultServerSpec{
					BackendStorage: backendEtcdWithSecret,
				},
			},
			nil,
			getSecret(),
			false,
			map[string]string{filepath.Base(util.VaultConfigFile): getConfigData(t, "", &backendEtcdWithSecret)},
		},
		{
			"backend inmem, expected error,extra config doesn't exist",
			api.VaultServer{
				ObjectMeta: getVaultObjectMeta(5),
				Spec: api.VaultServerSpec{
					BackendStorage: backendInmem,
					ConfigMapName:  "extra-config-123456",
				},
			},
			nil,
			nil,
			true,
			nil,
		},
	}

	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if test.extraConfigM != nil {
				createConfigMap(t, vaultCtrl.kubeClient, test.extraConfigM)
				defer deleteConfigMap(t, vaultCtrl.kubeClient, test.extraConfigM)
			}
			if test.extraSecret != nil {
				createSecret(t, vaultCtrl.kubeClient, test.extraSecret)
				defer deleteSecret(t, vaultCtrl.kubeClient, test.extraSecret)
			}

			err := vaultCtrl.prepareConfig(&test.vs)
			if test.exptErr {
				assert.NotNil(t, err)

			} else {
				if assert.Nil(t, err) {
					cm, err := vaultCtrl.kubeClient.CoreV1().ConfigMaps(test.vs.Namespace).Get(util.ConfigMapNameForVault(&test.vs), metav1.GetOptions{})
					if assert.Nil(t, err) {
						assert.Equal(t, util.ConfigMapNameForVault(&test.vs), cm.Name)
						assert.Equal(t, test.exptConfigMData, cm.Data)
						defer deleteConfigMap(t, vaultCtrl.kubeClient, cm)
					}
				}
			}
		})
	}
}

func TestPrepareVaultTLSSecrets(t *testing.T) {
	testData := []struct {
		name        string
		vs          *api.VaultServer
		extraSecret *corev1.Secret
		epectErr    bool
	}{
		{
			"no error, secret already exists",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      VaultTlsSecretName,
					Namespace: getVaultObjectMeta(1).Namespace,
				},
			},
			false,
		},
		{
			"no error, create secret successfully",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(2),
			},
			nil,
			false,
		},
	}

	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if test.extraSecret != nil {
				createSecret(t, vaultCtrl.kubeClient, test.extraSecret)
				defer deleteSecret(t, vaultCtrl.kubeClient, test.extraSecret)
			}

			err := vaultCtrl.prepareVaultTLSSecrets(test.vs)
			if test.epectErr {
				assert.NotNil(t, err)
				sr, err := vaultCtrl.kubeClient.CoreV1().Secrets(test.vs.Namespace).Get(VaultTlsSecretName, metav1.GetOptions{})
				defer deleteSecret(t, vaultCtrl.kubeClient, sr)
				assert.Nil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestReconcileVault(t *testing.T) {
	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
		recorder:   record.NewFakeRecorder(0),
		ctxCancels: map[string]context.CancelFunc{},
	}

	testData := []struct {
		name            string
		vs              *api.VaultServer
		extraDeployment *appsv1beta1.Deployment
		expectErr       bool
	}{
		{
			"create vault cluster, no error",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
				Spec: api.VaultServerSpec{
					Nodes: vaultTestNodes,
					BackendStorage: api.BackendStorageSpec{
						Inmem: true,
					},
				},
			},
			nil,
			false,
		},
		{
			"update vault replicas, no error",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(2),
				Spec: api.VaultServerSpec{
					Nodes:     vaultTestNodes,
					BaseImage: "vault",
					BackendStorage: api.BackendStorageSpec{
						Inmem: true,
					},
				},
			},
			getDeployment(t, getVaultObjectMeta(2), "vault:1.1.1", 1),
			false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if test.extraDeployment != nil {
				createDeployment(t, vaultCtrl.kubeClient, test.extraDeployment)
				defer deleteDeployment(t, vaultCtrl.kubeClient, test.extraDeployment)
			}

			// to ignore monitorAndUpdateStatus
			vaultCtrl.ctxCancels[test.vs.Name] = func() {}

			err := vaultCtrl.reconcileVault(test.vs)
			if test.expectErr {
				assert.NotNil(t, err, "error must be non-empty")
			} else {
				assert.Nil(t, err, "error must be nil")

				d, err := vaultCtrl.kubeClient.AppsV1beta1().Deployments(test.vs.Namespace).Get(test.vs.Name, metav1.GetOptions{})
				assert.Nil(t, err, "deployment for vaultserver should exist")
				assert.Equal(t, *d.Spec.Replicas, test.vs.Spec.Nodes)
			}
		})
	}
}

func TestDeployVault(t *testing.T) {
	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	testData := []struct {
		name            string
		vs              *api.VaultServer
		expectErr       bool
		extraDeployment *appsv1beta1.Deployment
		extraService    *corev1.Service
	}{
		{
			"vault deploy successful",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
				Spec: api.VaultServerSpec{
					BackendStorage: api.BackendStorageSpec{
						Inmem: true,
					},
				},
			},
			false,
			nil,
			nil,
		},
		{
			"vault deploy, deployment already exist",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(2),
				Spec:       api.VaultServerSpec{},
			},
			false,
			&appsv1beta1.Deployment{
				ObjectMeta: getVaultObjectMeta(2),
			},
			&corev1.Service{
				ObjectMeta: getVaultObjectMeta(2),
			},
		},
		{
			"vault deploy unsuccessful, service already exist",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(3),
				Spec:       api.VaultServerSpec{},
			},
			true,
			nil,
			&corev1.Service{
				ObjectMeta: getVaultObjectMeta(3),
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if test.extraDeployment != nil {
				createDeployment(t, vaultCtrl.kubeClient, test.extraDeployment)
				defer deleteDeployment(t, vaultCtrl.kubeClient, test.extraDeployment)
			}
			if test.extraService != nil {
				createService(t, vaultCtrl.kubeClient, test.extraService)
				defer deleteService(t, vaultCtrl.kubeClient, test.extraService)
			}

			err := vaultCtrl.DeployVault(test.vs)
			if test.expectErr {
				assert.NotNil(t, err, "error must be non-empty")
			} else {
				assert.Nil(t, err)
				_, err = vaultCtrl.kubeClient.AppsV1beta1().Deployments(test.vs.Namespace).Get(test.vs.Name, metav1.GetOptions{})
				assert.Nil(t, err, "deployment for vaultserver should exist")
				_, err = vaultCtrl.kubeClient.CoreV1().Services(test.vs.Namespace).Get(test.vs.Name, metav1.GetOptions{})
				assert.Nil(t, err, "service for vaultserver should exist")
			}
		})
	}
}

func TestSyncUpgrade(t *testing.T) {
	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	testData := []struct {
		name      string
		vs        *api.VaultServer
		d         *appsv1beta1.Deployment
		expectErr bool
	}{
		{
			"update image,no error",
			&api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
				Spec: api.VaultServerSpec{
					BaseImage: "vault",
					Version:   "0.1.0",
				},
			},
			getDeployment(t, getVaultObjectMeta(1), "vault:0.0.0"),
			false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if test.d != nil {
				createDeployment(t, vaultCtrl.kubeClient, test.d)
				defer deleteDeployment(t, vaultCtrl.kubeClient, test.d)
			}

			err := vaultCtrl.syncUpgrade(test.vs, test.d)
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCreateRoleAndRoleBinding(t *testing.T) {
	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	demoRole := rbac.Role{
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secret"},
				Verbs:     []string{"*"},
			},
		},
	}

	vs := &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "try",
			Namespace: "default",
		},
	}

	testData := []struct {
		testName           string
		preCreatedRole     []rbac.Role
		roles              []rbac.Role
		expectErr          bool
		expectRoles        []string
		expectRoleBindings []string
	}{
		{
			"create 2 rbac role and rolebinding",
			nil,
			[]rbac.Role{
				func(r *rbac.Role) rbac.Role { r.SetName("test1"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
				func(r *rbac.Role) rbac.Role { r.SetName("test2"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
			},
			false,
			[]string{"test1", "test2"},
			[]string{"test1", "test2"},
		},
		{
			"create 1 rbac role and rolebinding, but role already exists",
			[]rbac.Role{
				func(r *rbac.Role) rbac.Role { r.SetName("test3"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
			},
			[]rbac.Role{
				func(r *rbac.Role) rbac.Role { r.SetName("test3"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
			},
			false,
			[]string{"test3"},
			[]string{"test3"},
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			for _, r := range test.preCreatedRole {
				_, err := vaultCtrl.kubeClient.RbacV1().Roles(vs.Namespace).Create(&r)
				assert.Nil(t, err)
			}

			err := vaultCtrl.createRoleAndRoleBinding(vs, test.roles, "try")
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			for _, r := range test.expectRoles {
				_, err := vaultCtrl.kubeClient.RbacV1().Roles(vs.Namespace).Get(r, metav1.GetOptions{})
				assert.Nil(t, err, fmt.Sprintf("role(%s) should exists", r))
			}

			for _, rb := range test.expectRoleBindings {
				_, err := vaultCtrl.kubeClient.RbacV1().RoleBindings(vs.Namespace).Get(rb, metav1.GetOptions{})
				assert.Nil(t, err, fmt.Sprintf("rolebinding (%s) should exists", rb))
			}
		})
	}
}

func TestCreateSecret(t *testing.T) {
	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	demoSecret := corev1.Secret{
		Data: map[string][]byte{
			"test": []byte("secret"),
		},
	}

	vs := &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "try",
			Namespace: "default",
		},
	}

	testData := []struct {
		testName         string
		preCreatedSecret []corev1.Secret
		secrets          []corev1.Secret
		expectErr        bool
		expectSecrets    []string
	}{
		{
			"create 2 secret",
			nil,
			[]corev1.Secret{
				func(r *corev1.Secret) corev1.Secret { r.SetName("test1"); r.SetNamespace(vs.Namespace); return *r }(&demoSecret),
				func(r *corev1.Secret) corev1.Secret { r.SetName("test2"); r.SetNamespace(vs.Namespace); return *r }(&demoSecret),
			},
			false,
			[]string{"test1", "test2"},
		},
		{
			"create 1 secret, but secret already exist",
			[]corev1.Secret{
				func(r *corev1.Secret) corev1.Secret { r.SetName("test3"); r.SetNamespace(vs.Namespace); return *r }(&demoSecret),
			},
			[]corev1.Secret{
				func(r *corev1.Secret) corev1.Secret { r.SetName("test3"); r.SetNamespace(vs.Namespace); return *r }(&demoSecret),
			},
			false,
			[]string{"test3"},
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			for _, s := range test.preCreatedSecret {
				_, err := vaultCtrl.kubeClient.CoreV1().Secrets(vs.Namespace).Create(&s)
				assert.Nil(t, err)
			}

			err := vaultCtrl.createSecret(vs, test.secrets, "")
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			for _, s := range test.expectSecrets {
				_, err := vaultCtrl.kubeClient.CoreV1().Secrets(vs.Namespace).Get(s, metav1.GetOptions{})
				assert.Nil(t, err, fmt.Sprintf("secret(%s) should exists", s))
			}

		})
	}
}
