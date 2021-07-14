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

package controller

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/exporter"
	"kubevault.dev/operator/pkg/vault/storage"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (s *storageFake) Apply(pt *core.PodTemplateSpec) error {
	if s.ErrInApply {
		return fmt.Errorf("error")
	}
	return nil
}

type exporterFake struct{}

func (exp *exporterFake) Apply(pt *core.PodTemplateSpec, vs *api.VaultServer) error {
	return nil
}

func (exp *exporterFake) GetTelemetryConfig() (string, error) {
	return "", nil
}

type unsealerFake struct {
	ErrInApply bool
}

func (u *unsealerFake) Apply(pt *core.PodTemplateSpec) error {
	if u.ErrInApply {
		return fmt.Errorf("error")
	}
	return nil
}

func (u *unsealerFake) GetRBAC(prefix, namespace string) []rbacv1.Role {
	return nil
}

func getVaultObjectMeta(i int) metav1.ObjectMeta {
	suffix := strconv.Itoa(i)
	return metav1.ObjectMeta{
		Name:      vaultTestName + suffix,
		Namespace: vaultTestNamespace + suffix,
	}
}

func getConfigData(extraConfig string, storageCfg string, exptrCfg string) string {
	cfg := util.GetListenerConfig("/etc/vault/tls/server/", false)
	if len(extraConfig) != 0 {
		cfg = fmt.Sprintf("%s\n%s", cfg, extraConfig)
	}
	uiCfg := "ui = true"
	cfg = fmt.Sprintf("%s\n%s\n%s\n%s", cfg, uiCfg, storageCfg, exptrCfg)
	return cfg
}

func TestGetConfig(t *testing.T) {
	var (
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
		exporter        exporter.Exporter
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
			exporter:        &exporterFake{},
			exptErr:         false,
			exptConfigMData: map[string]string{filepath.Base(util.VaultConfigFile): getConfigData("", storageCfg, "")},
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
			exporter:        &exporterFake{},
			exptErr:         true,
			exptConfigMData: nil,
		},
	}

	for idx := range testData {
		test := testData[idx]
		t.Run(test.name, func(t *testing.T) {
			v := vaultSrv{
				kubeClient: kfake.NewSimpleClientset(),
				vs:         &test.vs,
				strg:       test.storage,
				exprtr:     test.exporter,
			}
			cm, err := v.GetConfig()
			if test.exptErr {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, test.vs.ConfigSecretName(), cm.Name)
					mss := make(map[string]string, len(cm.Data))
					for k, v := range cm.Data {
						mss[k] = string(v)
					}
					assert.Equal(t, test.exptConfigMData, mss)
				}
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
		pt        *core.PodTemplateSpec
		unslr     *unsealerFake
		strg      *storageFake
		expectErr bool
	}{
		{
			name: "no error",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			pt:        &core.PodTemplateSpec{},
			unslr:     &ufake,
			strg:      &sfake,
			expectErr: false,
		},
		{
			name: "error for storage",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			pt:        &core.PodTemplateSpec{},
			unslr:     &ufake,
			strg:      func(s storageFake) *storageFake { s.ErrInApply = true; return &s }(sfake),
			expectErr: true,
		},
		{
			name: "error for unsealer",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},
			pt:        &core.PodTemplateSpec{},
			unslr:     func(u unsealerFake) *unsealerFake { u.ErrInApply = true; return &u }(ufake),
			strg:      &sfake,
			expectErr: true,
		},
	}

	for idx := range testData {
		test := testData[idx]
		t.Run(test.name, func(t *testing.T) {
			v := vaultSrv{
				kubeClient: kfake.NewSimpleClientset(),
				vs:         test.vs,
				strg:       test.strg,
				exprtr:     &exporterFake{},
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
