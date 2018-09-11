package controller

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/storage"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"
)

const (
	vaultTestName      = "ex-vault-unit-test"
	vaultTestNamespace = "ns-vault-test"
)

type storageFake struct {
	config          string
	ErrInGetStrgCfg bool
	ErrInApply      bool
}

func (s *storageFake) GetStorageConfig() (string, error) {
	if s.ErrInGetStrgCfg {
		return "", fmt.Errorf("error")
	}
	return s.config, nil
}

func (s *storageFake) Apply(pt *corev1.PodTemplateSpec) error {
	if s.ErrInApply {
		return fmt.Errorf("error")
	}
	return nil
}

type unsealerFake struct {
	ErrInApply bool
}

func (u *unsealerFake) Apply(pt *corev1.PodTemplateSpec) error {
	if u.ErrInApply {
		return fmt.Errorf("error")
	}
	return nil
}

func (u *unsealerFake) GetRBAC(namespace string) []rbacv1.Role {
	return nil
}

func getVaultObjectMeta(i int) metav1.ObjectMeta {
	suffix := strconv.Itoa(i)
	return metav1.ObjectMeta{
		Name:      vaultTestName + suffix,
		Namespace: vaultTestNamespace + suffix,
	}
}

func getConfigData(t *testing.T, extraConfig string, storageCfg string) string {
	cfg := util.GetListenerConfig()
	if len(extraConfig) != 0 {
		cfg = fmt.Sprintf("%s\n%s", cfg, extraConfig)
	}
	cfg = fmt.Sprintf("%s\n%s", cfg, storageCfg)
	return cfg
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

func TestGetConfig(t *testing.T) {
	var (
		extraCfg = `
telemetry {
  statsite_address = "statsite.test.local:8125"
}
`
		storageCfg = `
storage "test"{
	hi = "hello"
	one = "two"
}
`
	)

	testData := []struct {
		name            string
		vs              api.VaultServer
		storage         storage.Storage
		extraConfigM    *corev1.ConfigMap
		exptErr         bool
		exptConfigMData map[string]string
	}{
		{
			name: "with no extra config",
			vs: api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			storage: &storageFake{
				config:          storageCfg,
				ErrInGetStrgCfg: false,
			},
			extraConfigM:    nil,
			exptErr:         false,
			exptConfigMData: map[string]string{filepath.Base(util.VaultConfigFile): getConfigData(t, "", storageCfg)},
		},
		{
			name: "with extra config",
			vs: api.VaultServer{
				ObjectMeta: getVaultObjectMeta(2),
				Spec: api.VaultServerSpec{
					ConfigMapName: getVaultObjectMeta(2).Name + "-config",
				},
			},
			storage: &storageFake{
				config:          storageCfg,
				ErrInGetStrgCfg: false,
			},
			extraConfigM:    getConfigMap(getVaultObjectMeta(2), extraCfg),
			exptErr:         false,
			exptConfigMData: map[string]string{filepath.Base(util.VaultConfigFile): getConfigData(t, extraCfg, storageCfg)},
		},
		{
			name: "expected error,configMap with extra config doesn't exist",
			vs: api.VaultServer{
				ObjectMeta: getVaultObjectMeta(5),
				Spec: api.VaultServerSpec{
					ConfigMapName: getVaultObjectMeta(5).Name + "-config",
				},
			},
			storage: &storageFake{
				config:          storageCfg,
				ErrInGetStrgCfg: false,
			},
			extraConfigM:    nil,
			exptErr:         true,
			exptConfigMData: nil,
		},
		{
			name: "expected error, error when getting storage config",
			vs: api.VaultServer{
				ObjectMeta: getVaultObjectMeta(5),
			},
			storage: &storageFake{
				config:          storageCfg,
				ErrInGetStrgCfg: true,
			},
			extraConfigM:    nil,
			exptErr:         true,
			exptConfigMData: nil,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			v := vaultSrv{
				kubeClient: kfake.NewSimpleClientset(),
				vs:         &test.vs,
				strg:       test.storage,
			}
			if test.extraConfigM != nil {
				createConfigMap(t, v.kubeClient, test.extraConfigM)
				defer deleteConfigMap(t, v.kubeClient, test.extraConfigM)
			}
			cm, err := v.GetConfig()
			if test.exptErr {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, test.vs.ConfigMapName(), cm.Name)
					assert.Equal(t, test.exptConfigMData, cm.Data)
				}
			}
		})
	}
}

func TestGetServerTLS(t *testing.T) {
	testData := []struct {
		name        string
		vs          *api.VaultServer
		extraSecret *corev1.Secret
		expectErr   bool
	}{
		{
			name: "no error, secret already exists",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			extraSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: api.VaultServer{
						ObjectMeta: getVaultObjectMeta(1),
					}.TLSSecretName(),
					Namespace: getVaultObjectMeta(1).Namespace,
				},
			},
			expectErr: false,
		},
		{
			name: "no error, user provided secret",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
				Spec: api.VaultServerSpec{
					TLS: &api.TLSPolicy{
						TLSSecret: "vault-tls-cred",
					},
				},
			},
			extraSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vault-tls-cred",
					Namespace: getVaultObjectMeta(1).Namespace,
				},
			},
			expectErr: false,
		},
		{
			name: "expect error, user provided secret",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
				Spec: api.VaultServerSpec{
					TLS: &api.TLSPolicy{
						TLSSecret: "vault-tls-cred",
					},
				},
			},
			extraSecret: nil,
			expectErr:   true,
		},
		{
			name: "no error, create secret successfully",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(2),
			},
			extraSecret: nil,
			expectErr:   false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			v := vaultSrv{
				kubeClient: kfake.NewSimpleClientset(),
				vs:         test.vs,
			}

			if test.extraSecret != nil {
				createSecret(t, v.kubeClient, test.extraSecret)
				defer deleteSecret(t, v.kubeClient, test.extraSecret)
			}

			_, err := v.GetServerTLS()
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestApply(t *testing.T) {
	ufake := unsealerFake{}
	sfake := storageFake{}
	testData := []struct {
		name      string
		vs        *api.VaultServer
		pt        *corev1.PodTemplateSpec
		unslr     *unsealerFake
		strg      *storageFake
		expectErr bool
	}{
		{
			name: "no error",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			pt:        &corev1.PodTemplateSpec{},
			unslr:     &ufake,
			strg:      &sfake,
			expectErr: false,
		},
		{
			name: "error for storage",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			pt:        &corev1.PodTemplateSpec{},
			unslr:     &ufake,
			strg:      func(s storageFake) *storageFake { s.ErrInApply = true; return &s }(sfake),
			expectErr: true,
		},
		{
			name: "error for unsealer",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			pt:        &corev1.PodTemplateSpec{},
			unslr:     func(u unsealerFake) *unsealerFake { u.ErrInApply = true; return &u }(ufake),
			strg:      &sfake,
			expectErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			v := vaultSrv{
				kubeClient: kfake.NewSimpleClientset(),
				vs:         test.vs,
				strg:       test.strg,
				unslr:      test.unslr,
			}

			err := v.Apply(test.pt)
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
