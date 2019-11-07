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

package admission

import (
	"testing"

	catalog "kubevault.dev/operator/apis/catalog/v1alpha1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	extfake "kubevault.dev/operator/client/clientset/versioned/fake"
	clientsetscheme "kubevault.dev/operator/client/clientset/versioned/scheme"

	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	meta_util "kmodules.xyz/client-go/meta"
)

const namespace = "test-ns"

var (
	vsVersion = catalog.VaultServerVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "1.11.1",
			Namespace: namespace,
		},
	}
	vs = api.VaultServer{
		TypeMeta: metav1.TypeMeta{
			Kind:       api.ResourceKindVaultServer,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vs",
			Namespace: namespace,
		},
		Spec: api.VaultServerSpec{
			Version:  "1.11.1",
			Nodes:    1,
			TLS:      nil,
			Backend:  api.BackendStorageSpec{},
			Unsealer: nil,
		},
	}
	unslr = api.UnsealerSpec{
		SecretShares:    5,
		SecretThreshold: 3,
		Mode:            api.ModeSpec{},
	}
)

func init() {
	utilruntime.Must(clientsetscheme.AddToScheme(scheme.Scheme))
}

func TestVaultServerValidator_Admit(t *testing.T) {
	cases := []struct {
		testName  string
		operation admission.Operation
		object    api.VaultServer
		oldObject api.VaultServer
		allowed   bool
	}{
		{
			testName:  "Create valid VaultServer, operation allowed",
			operation: admission.Create,
			object:    validVaultServer(),
			oldObject: vs,
			allowed:   true,
		},
		{
			testName:  "Create invalid VaultServer, operation not allowed",
			operation: admission.Create,
			object:    invalidVaultServer(),
			oldObject: vs,
			allowed:   false,
		},
		{
			testName:  "Update VaultServer nodes, operation allowed",
			operation: admission.Update,
			object:    func() api.VaultServer { v := validVaultServer(); v.Spec.Nodes = 10; return v }(),
			oldObject: validVaultServer(),
			allowed:   true,
		},
		{
			testName:  "Update VaultServer unsealer, operation not allowed",
			operation: admission.Update,
			object:    func() api.VaultServer { u := unsealerWithKubernetes(); return vaultServerWiitUnsealer(&u) }(),
			oldObject: vaultServerWiitUnsealer(nil),
			allowed:   false,
		},
		{
			testName:  "Delete VaultServer, operation allowed",
			operation: admission.Delete,
			object:    validVaultServer(),
			oldObject: validVaultServer(),
			allowed:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := &VaultServerValidator{
				client:      kfake.NewSimpleClientset(),
				extClient:   extfake.NewSimpleClientset(&vsVersion),
				initialized: true,
			}

			rawObj, err := meta_util.MarshalToJson(&c.object, api.SchemeGroupVersion)
			if !assert.Nil(t, err, "VaultServer marshal to json failed") {
				return
			}
			rawOldObj, err := meta_util.MarshalToJson(&c.oldObject, api.SchemeGroupVersion)
			if !assert.Nil(t, err, "VaultServer marshal to json failed") {
				return
			}

			req := &admission.AdmissionRequest{
				Kind: metav1.GroupVersionKind{
					Group:   api.SchemeGroupVersion.Group,
					Kind:    api.ResourceKindVaultServer,
					Version: api.SchemeGroupVersion.Version,
				},
				Operation: c.operation,

				Object: runtime.RawExtension{
					Raw: rawObj,
				},
				OldObject: runtime.RawExtension{
					Raw: rawOldObj,
				},
			}

			resp := validator.Admit(req)
			assert.Equal(t, c.allowed, resp.Allowed, "admission response")
		})
	}
}

func TestValidateVaultServer(t *testing.T) {
	cases := []struct {
		testName    string
		vs          *api.VaultServer
		extraSecret []core.Secret
		expectErr   bool
	}{
		{
			testName:    "spec.version is missing, expect error",
			vs:          func() *api.VaultServer { v := vs; v.Spec.Version = "not-exist"; return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "spec.nodes is invalid, expect error",
			vs:          func() *api.VaultServer { v := vs; v.Spec.Nodes = 0; return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "spec.tls.tlsSecret is empty, expect error",
			vs:          func() *api.VaultServer { v := vs; v.Spec.TLS = &api.TLSPolicy{TLSSecret: ""}; return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "number of backend is not specified, expect error",
			vs:          func() *api.VaultServer { v := vs; return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "number of backend is more than 1, expect error",
			vs: func() *api.VaultServer {
				v := vs
				v.Spec.Backend = api.BackendStorageSpec{Inmem: &api.InmemSpec{}, File: &api.FileSpec{}}
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.etcd , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithEtcd(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.mySQL , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithMySQL(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.postgreSQL , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithPostgres(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.gcs , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithGcs(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.s3 , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithS3(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.azure , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithAzure(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.dynamoDB , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithDynamoDB(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error for spec.backend.swift , expect error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithSwift(); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "using inmem, expect no error",
			vs: func() *api.VaultServer {
				v := vs
				v.Spec.Backend = api.BackendStorageSpec{Inmem: &api.InmemSpec{}}
				return &v
			}(),
			extraSecret: nil,
			expectErr:   false,
		},
		{
			testName: "using file, expect no error",
			vs: func() *api.VaultServer {
				v := vs
				v.Spec.Backend = api.BackendStorageSpec{File: &api.FileSpec{}}
				return &v
			}(),
			extraSecret: nil,
			expectErr:   false,
		},
		{
			testName:    "using etcd, expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithEtcd(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithEtcd(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using mySQL , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithMySQL(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithMySQL(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using postgreSQL , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithPostgres(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithPostgres(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using gcs , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithGcs(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithGcs(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using s3 , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithS3(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithS3(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using azure , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithAzure(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithAzure(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using dynamoDB , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithDynamoDB(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithDynamoDB(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using swift , expect no error",
			vs:          func() *api.VaultServer { v, _ := vaultServerWithSwift(); return &v }(),
			extraSecret: func() []core.Secret { _, s := vaultServerWithSwift(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "spec.unsealer.secretShares is zero, expect error",
			vs:          func() *api.VaultServer { u := unslr; u.SecretShares = 0; v := vaultServerWiitUnsealer(&u); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "spec.unsealer.secretThreshold is zero, expect error",
			vs: func() *api.VaultServer {
				u := unslr
				u.SecretThreshold = 0
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "spec.unsealer.secretThreshold > spec.unsealer.secretShares, expect error",
			vs: func() *api.VaultServer {
				u := unslr
				u.SecretThreshold = u.SecretShares + 1
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "spec.unsealer.insecureTLS is false and spec.unsealer.vaultCASecret is empty, expect error",
			vs: func() *api.VaultServer {
				u := unslr
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "spec.unsealer.mode is empty, expect error",
			vs:          func() *api.VaultServer { u := unslr; v := vaultServerWiitUnsealer(&u); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "specifies more than one modes in spec.unsealer.mode, expect error",
			vs: func() *api.VaultServer {
				u := unsealerWithKubernetes()
				u.Mode.GoogleKmsGcs = &api.GoogleKmsGcsSpec{}
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "secret validation error in GoogleKmsGcs unsealer, expect error",
			vs: func() *api.VaultServer {
				u, _ := unsealerWithGoogleKmsGcs()
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "secret validation error in AwsKmsSsm unsealer, expect error",
			vs:          func() *api.VaultServer { u, _ := unsealerWithAwsKmsSsm(); v := vaultServerWiitUnsealer(&u); return &v }(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName: "secret validation error in AzureKeyVault unsealer, expect error",
			vs: func() *api.VaultServer {
				u, _ := unsealerWithAzureKeyVault()
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: nil,
			expectErr:   true,
		},
		{
			testName:    "using kubernetes secret unsealer, expect no error",
			vs:          func() *api.VaultServer { u := unsealerWithKubernetes(); v := vaultServerWiitUnsealer(&u); return &v }(),
			extraSecret: nil,
			expectErr:   false,
		},
		{
			testName: "using  GoogleKmsGcs unsealer, expect no error",
			vs: func() *api.VaultServer {
				u, _ := unsealerWithGoogleKmsGcs()
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: func() []core.Secret { _, s := unsealerWithGoogleKmsGcs(); return s }(),
			expectErr:   false,
		},
		{
			testName:    "using  AwsKmsSsm unsealer, expect no error",
			vs:          func() *api.VaultServer { u, _ := unsealerWithAwsKmsSsm(); v := vaultServerWiitUnsealer(&u); return &v }(),
			extraSecret: func() []core.Secret { _, s := unsealerWithAwsKmsSsm(); return s }(),
			expectErr:   false,
		},
		{
			testName: "using  AzureKeyVault unsealer, expect no error",
			vs: func() *api.VaultServer {
				u, _ := unsealerWithAzureKeyVault()
				v := vaultServerWiitUnsealer(&u)
				return &v
			}(),
			extraSecret: func() []core.Secret { _, s := unsealerWithAzureKeyVault(); return s }(),
			expectErr:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			kc := kfake.NewSimpleClientset()
			for _, sr := range c.extraSecret {
				_, err := kc.CoreV1().Secrets(sr.Namespace).Create(&sr)
				assert.Nil(t, err, "create secret error should be nil")
			}

			extC := extfake.NewSimpleClientset(&vsVersion)

			err := ValidateVaultServer(kc, extC, c.vs)
			if c.expectErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err, "error should be nil")
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	cases := []struct {
		testName    string
		object      api.VaultServer
		oldObject   api.VaultServer
		expectedErr bool
	}{
		{
			testName:    "update nodes, expect no error",
			object:      func() api.VaultServer { v := validVaultServer(); v.Spec.Nodes = 10; return v }(),
			oldObject:   validVaultServer(),
			expectedErr: false,
		},
		{
			testName: "update spec.unsealer.secretShares, expect error",
			object: func() api.VaultServer {
				u := unsealerWithKubernetes()
				u.SecretShares = 100
				return vaultServerWiitUnsealer(&u)
			}(),
			oldObject:   func() api.VaultServer { u := unsealerWithKubernetes(); return vaultServerWiitUnsealer(&u) }(),
			expectedErr: true,
		},
		{
			testName:    "update spec.backend.file.path , expect error",
			object:      func() api.VaultServer { v := vaultServerWithFile(); v.Spec.Backend.File.Path = "/new"; return v }(),
			oldObject:   func() api.VaultServer { v := vaultServerWithFile(); ; return v }(),
			expectedErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			err := validateUpdate(&c.object, &c.oldObject)
			if c.expectedErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err, "error should be nil")
			}
		})
	}
}

func TestValidateSecret(t *testing.T) {
	sr := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-test-1",
			Namespace: "test",
		},
		Data: map[string][]byte{
			"key1": []byte("value1"),
			"key2": []byte(""),
		},
	}

	cases := []struct {
		testName     string
		secret       *core.Secret
		requiredKeys []string
		expectErr    bool
	}{
		{
			testName: "secret exist, all required keys are present, expect no error",
			secret:   sr,
			requiredKeys: []string{
				"key1",
			},
			expectErr: false,
		},
		{
			testName: "secret exist, all required keys aren't present, expect error",
			secret:   sr,
			requiredKeys: []string{
				"key1",
				"key3",
			},
			expectErr: true,
		},
		{
			testName: "secret exist, all required keys are present but one them has empty value, expect error",
			secret:   sr,
			requiredKeys: []string{
				"key1",
				"key2",
			},
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			kc := kfake.NewSimpleClientset(c.secret)
			err := validateSecret(kc, c.secret.Name, c.secret.Namespace, c.requiredKeys)
			if c.expectErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err, "error should be nil")
			}
		})
	}
}

func vaultServerWithFile() api.VaultServer {
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		File: &api.FileSpec{
			Path: "/etc/vault",
		},
	}
	return v
}

func vaultServerWithEtcd() (api.VaultServer, []core.Secret) {
	etcd := &api.EtcdSpec{
		CredentialSecretName: "etcd-cred",
		TLSSecretName:        "etcd-tls",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		Etcd: etcd,
	}
	extraSr := []core.Secret{
		getSecret(etcd.CredentialSecretName, []string{
			"username",
			"password",
		}),
		getSecret(etcd.TLSSecretName, []string{
			"ca.crt",
			"client.crt",
			"client.key",
		}),
	}
	return v, extraSr
}

func vaultServerWithPostgres() (api.VaultServer, []core.Secret) {
	pg := &api.PostgreSQLSpec{
		ConnectionUrlSecret: "pg-con",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		PostgreSQL: pg,
	}
	extraSr := []core.Secret{
		getSecret(pg.ConnectionUrlSecret, []string{
			"connection_url",
		}),
	}
	return v, extraSr
}

func vaultServerWithMySQL() (api.VaultServer, []core.Secret) {
	my := &api.MySQLSpec{
		UserCredentialSecret: "my-cred",
		TLSCASecret:          "my-tls",
	}

	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		MySQL: my,
	}

	extraSr := []core.Secret{
		getSecret(my.UserCredentialSecret, []string{
			"username",
			"password",
		}),
		getSecret(my.TLSCASecret, []string{
			"tls_ca_file",
		}),
	}
	return v, extraSr
}

func vaultServerWithGcs() (api.VaultServer, []core.Secret) {
	gcs := &api.GcsSpec{
		CredentialSecret: "gcs-sa",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		Gcs: gcs,
	}
	extraSr := []core.Secret{
		getSecret(gcs.CredentialSecret, []string{
			"sa.json",
		}),
	}
	return v, extraSr
}

func vaultServerWithS3() (api.VaultServer, []core.Secret) {
	s3 := &api.S3Spec{
		CredentialSecret:   "s3-cred",
		SessionTokenSecret: "s3-token",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		S3: s3,
	}
	extraSr := []core.Secret{
		getSecret(s3.CredentialSecret, []string{
			"access_key",
			"secret_key",
		}),
		getSecret(s3.SessionTokenSecret, []string{
			"session_token",
		}),
	}
	return v, extraSr
}

func vaultServerWithAzure() (api.VaultServer, []core.Secret) {
	az := &api.AzureSpec{
		AccountKeySecret: "az-ac",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		Azure: az,
	}
	extraSr := []core.Secret{
		getSecret(az.AccountKeySecret, []string{
			"account_key",
		}),
	}
	return v, extraSr
}

func vaultServerWithDynamoDB() (api.VaultServer, []core.Secret) {
	db := &api.DynamoDBSpec{
		CredentialSecret:   "db-cred",
		SessionTokenSecret: "db-token",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		DynamoDB: db,
	}
	extraSr := []core.Secret{
		getSecret(db.CredentialSecret, []string{
			"access_key",
			"secret_key",
		}),
		getSecret(db.SessionTokenSecret, []string{
			"session_token",
		}),
	}
	return v, extraSr
}

func vaultServerWithSwift() (api.VaultServer, []core.Secret) {
	sw := &api.SwiftSpec{
		CredentialSecret: "sw-cred",
		AuthTokenSecret:  "sw-token",
	}
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		Swift: sw,
	}
	extraSr := []core.Secret{
		getSecret(sw.CredentialSecret, []string{
			"username",
			"password",
		}),
		getSecret(sw.AuthTokenSecret, []string{
			"auth_token",
		}),
	}
	return v, extraSr
}

func vaultServerWiitUnsealer(u *api.UnsealerSpec) api.VaultServer {
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		Inmem: &api.InmemSpec{},
	}
	v.Spec.Unsealer = u
	return v
}

func unsealerWithKubernetes() api.UnsealerSpec {
	u := unslr
	u.Mode = api.ModeSpec{
		KubernetesSecret: &api.KubernetesSecretSpec{},
	}
	return u
}

func unsealerWithGoogleKmsGcs() (api.UnsealerSpec, []core.Secret) {
	u := unslr
	u.Mode = api.ModeSpec{
		GoogleKmsGcs: &api.GoogleKmsGcsSpec{
			CredentialSecret: "g-cred",
		},
	}
	extraSr := []core.Secret{
		getSecret("g-cred", []string{
			"sa.json",
		}),
	}
	return u, extraSr
}

func unsealerWithAwsKmsSsm() (api.UnsealerSpec, []core.Secret) {
	u := unslr
	u.Mode = api.ModeSpec{
		AwsKmsSsm: &api.AwsKmsSsmSpec{
			CredentialSecret: "aws-cred",
		},
	}
	extraSr := []core.Secret{
		getSecret("aws-cred", []string{
			"access_key",
			"secret_key",
		}),
	}
	return u, extraSr
}

func unsealerWithAzureKeyVault() (api.UnsealerSpec, []core.Secret) {
	u := unslr
	u.Mode = api.ModeSpec{
		AzureKeyVault: &api.AzureKeyVault{
			AADClientSecret:  "az-add",
			ClientCertSecret: "az-client",
		},
	}
	extraSr := []core.Secret{
		getSecret("az-add", []string{
			"client-id",
			"client-secret",
		}), getSecret("az-client", []string{
			"client-cert",
			"client-cert-password",
		}),
	}
	return u, extraSr
}

func validVaultServer() api.VaultServer {
	v := vs
	v.Spec.Backend = api.BackendStorageSpec{
		Inmem: &api.InmemSpec{},
	}
	return v
}

func invalidVaultServer() api.VaultServer {
	return vs
}

func getSecret(name string, keys []string) core.Secret {
	sr := core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{},
	}
	for _, k := range keys {
		sr.Data[k] = []byte("value")
	}
	return sr
}
