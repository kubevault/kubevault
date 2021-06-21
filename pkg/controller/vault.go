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
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math"
	"math/big"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	conapi "kubevault.dev/apimachinery/apis"
	capi "kubevault.dev/apimachinery/apis/catalog/v1alpha1"
	"kubevault.dev/apimachinery/apis/kubevault"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	cs "kubevault.dev/apimachinery/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault/exporter"
	"kubevault.dev/operator/pkg/vault/storage"
	"kubevault.dev/operator/pkg/vault/unsealer"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	"gomodules.xyz/cert"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	certlib "kubedb.dev/elasticsearch/pkg/lib/cert"
)

const (
	EnvVaultAPIAddr         = "VAULT_API_ADDR"
	EnvVaultClusterAddr     = "VAULT_CLUSTER_ADDR"
	EnvVaultCACert          = "VAULT_CACERT"
	VaultClientPort         = 8200
	VaultClusterPort        = 8201
	vaultTLSAssetVolumeName = "vault-tls-secret"
)

var (
	// ensure that s/a token is readable xref: https://issues.k8s.io/70679
	defaultFsGroup int64 = 65535
)

type Vault interface {
	EnsureCA() error
	GetCABundle() ([]byte, error)
	EnsureServerTLS() error
	EnsureClientTLS() error
	EnsureStorageTLS() error
	GetConfig() (*core.Secret, error)
	Apply(pt *core.PodTemplateSpec) error
	GetService() *core.Service
	GetGoverningService() *core.Service
	GetStatefulSet(serviceName string, pt *core.PodTemplateSpec, vcts []core.PersistentVolumeClaim) *apps.StatefulSet
	GetServiceAccounts() []core.ServiceAccount
	GetRBACRolesAndRoleBindings() ([]rbac.Role, []rbac.RoleBinding)
	GetRBACClusterRoleBinding() rbac.ClusterRoleBinding
	GetPodTemplate(c core.Container, saName string) *core.PodTemplateSpec
	GetContainer() core.Container
}

type vaultSrv struct {
	vs         *api.VaultServer
	strg       storage.Storage
	unslr      unsealer.Unsealer
	exprtr     exporter.Exporter
	kubeClient kubernetes.Interface
	config     capi.VaultServerVersionVault
}

func NewVault(vs *api.VaultServer, config *rest.Config, kc kubernetes.Interface, vc cs.Interface) (Vault, error) {
	version, err := vc.CatalogV1alpha1().VaultServerVersions().Get(context.TODO(), string(vs.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault server version")
	}

	// it is required to have storage
	strg, err := storage.NewStorage(kc, vs)
	if err != nil {
		return nil, err
	}

	// it is *not* required to have unsealer
	unslr, err := unsealer.NewUnsealerService(config, vs, version)
	if err != nil {
		return nil, err
	}

	exprtr, err := exporter.NewExporter(version, kc)
	if err != nil {
		return nil, err
	}

	return &vaultSrv{
		vs:         vs,
		strg:       strg,
		unslr:      unslr,
		exprtr:     exprtr,
		kubeClient: kc,
		config:     version.Spec.Vault,
	}, nil
}

func (v *vaultSrv) EnsureCA() error {
	// caSecretName
	//  if secret exist:
	//		validate
	//  else
	//	 	create CA
	//		create CA k8s Secret
	//	endif
	// return
	if v.vs.Spec.TLS == nil {
		return nil
	}
	// Certificates are managed by cert-manager
	if v.vs.Spec.TLS.IssuerRef != nil {
		return nil
	}

	caSecretName := v.vs.GetCertSecretName(string(api.VaultCACert))
	caSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), caSecretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			// Create new CA
			// Create k8s secret
			cfg := cert.Config{
				CommonName:   v.vs.GetCertificateCN(api.VaultCACert),
				Organization: []string{kubevault.GroupName},
			}

			caKey, err := cert.NewPrivateKey()
			if err != nil {
				return errors.New("failed to generate key for CA certificate")
			}

			caCert, err := cert.NewSelfSignedCACert(cfg, caKey)
			if err != nil {
				return errors.New("failed to generate CA certificate")
			}

			caKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(caKey)
			if err != nil {
				return errors.Wrap(err, "failed to encode private key")
			}

			caCertByte := cert.EncodeCertPEM(caCert)

			_, _, err = core_util.CreateOrPatchSecret(context.TODO(), v.kubeClient,
				metav1.ObjectMeta{
					Name:      v.vs.GetCertSecretName(string(api.VaultCACert)),
					Namespace: v.vs.Namespace,
				},
				func(in *core.Secret) *core.Secret {
					in.Labels = v.vs.OffshootLabels()
					core_util.EnsureOwnerReference(&in.ObjectMeta, metav1.NewControllerRef(v.vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
					in.Data = map[string][]byte{
						core.TLSCertKey:       caCertByte,
						core.TLSPrivateKeyKey: caKeyByte,
					}
					in.Type = core.SecretTypeTLS
					return in
				}, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to create ca k8s secret")
			}

		}
		return errors.Wrap(err, "failed to get secret")
	}
	// validate keys
	if value, exist := caSecret.Data[core.TLSCertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.crt", caSecret.Namespace, caSecret.Name)
	}

	if value, exist := caSecret.Data[core.TLSPrivateKeyKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.key", caSecret.Namespace, caSecret.Name)
	}

	return nil
}

func (v *vaultSrv) GetCABundle() ([]byte, error) {
	if v.vs.Spec.TLS == nil {
		return nil, errors.New("tls is disabled")
	}
	// Get Server secret
	sSecretName := v.vs.GetCertSecretName(string(api.VaultServerCert))
	sSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), sSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get server cert secret")
	}
	if value, exist := sSecret.Data[conapi.TLSCACertKey]; !exist || len(value) == 0 {
		return nil, errors.Errorf("%s/%s secret is missing ca.crt", sSecret.Namespace, sSecret.Name)
	} else {
		return value, nil
	}
}

func (v *vaultSrv) EnsureServerTLS() error {
	if v.vs.Spec.TLS == nil {
		return nil
	}
	// If Certificate is managed by cert-manager
	if v.vs.Spec.TLS.IssuerRef != nil {
		return nil
	}

	sSecretName := v.vs.GetCertSecretName(string(api.VaultServerCert))
	sSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), sSecretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			// Create new server TLS
			caSecretName := v.vs.GetCertSecretName(string(api.VaultCACert))
			caSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), caSecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// validate tls.crt
			caCertByte, exist := caSecret.Data[core.TLSCertKey]
			if !exist {
				return errors.Errorf("%s/%s does not contain tls.crt", caSecret.Namespace, caSecret.Name)
			}
			caCerts, err := cert.ParseCertsPEM(caCertByte)
			if err != nil {
				return err
			}
			if len(caCerts) == 0 {
				return errors.New("tls.crt is empty")
			}

			// validate tls.key
			caKeyByte, exist := caSecret.Data[core.TLSPrivateKeyKey]
			if !exist {
				return errors.Errorf("%s/%s does not contain tls.key", caSecret.Namespace, caSecret.Name)
			}

			caKeyInterface, err := cert.ParsePrivateKeyPEM(caKeyByte)
			if err != nil {
				return err
			}
			caKey, valid := caKeyInterface.(*rsa.PrivateKey)
			if !valid {
				klog.Infof("%v", caKey)
				return errors.New("Unsupported private key type")
			}

			cfg := cert.Config{
				CommonName:   v.vs.GetCertificateCN(api.VaultServerCert),
				Organization: []string{kubevault.GroupName},
				AltNames: cert.AltNames{
					DNSNames: []string{
						"localhost",
						fmt.Sprintf("*.%s.pod", v.vs.Namespace),
						fmt.Sprintf("%s.%s.svc", v.vs.ServiceName(api.VaultServerServiceVault), v.vs.Namespace),
						fmt.Sprintf("*.%s.%s.svc", v.vs.ServiceName(api.VaultServerServiceInternal), v.vs.Namespace),
						fmt.Sprintf("%s.%s.svc", v.vs.ServiceName(api.VaultServerServiceInternal), v.vs.Namespace),
					},
					IPs: []net.IP{
						net.ParseIP("127.0.0.1"),
					},
				},
				Usages: []x509.ExtKeyUsage{
					x509.ExtKeyUsageServerAuth,
				},
			}
			sKey, err := cert.NewPrivateKey()
			if err != nil {
				return errors.New("failed to generate key for server certificate")
			}
			sKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(sKey)
			if err != nil {
				return err
			}

			sCertificate, err := newSignedCert(cfg, sKey, caCerts[0], caKey)
			if err != nil {
				return errors.New("failed to sign server certificate")
			}
			sCertByte := cert.EncodeCertPEM(sCertificate)

			_, _, err = core_util.CreateOrPatchSecret(context.TODO(), v.kubeClient,
				metav1.ObjectMeta{
					Name:      v.vs.GetCertSecretName(string(api.VaultServerCert)),
					Namespace: v.vs.Namespace,
				},
				func(in *core.Secret) *core.Secret {
					in.Labels = v.vs.OffshootLabels()
					core_util.EnsureOwnerReference(&in.ObjectMeta, metav1.NewControllerRef(v.vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
					in.Data = map[string][]byte{
						core.TLSCertKey:       sCertByte,
						core.TLSPrivateKeyKey: sKeyByte,
						conapi.TLSCACertKey:   caCertByte,
					}
					in.Type = core.SecretTypeTLS
					return in
				}, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to create server k8s secret")
			}

		}
		return err
	}

	// validate keys
	if value, exist := sSecret.Data[conapi.TLSCACertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain ca.crt", sSecret.Namespace, sSecret.Name)
	}

	if value, exist := sSecret.Data[core.TLSCertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.crt", sSecret.Namespace, sSecret.Name)
	}

	if value, exist := sSecret.Data[core.TLSPrivateKeyKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.key", sSecret.Namespace, sSecret.Name)
	}

	return nil
}

func (v *vaultSrv) EnsureClientTLS() error {
	if v.vs.Spec.TLS == nil {
		return nil
	}
	// If Certificate is managed by cert-manager
	if v.vs.Spec.TLS.IssuerRef != nil {
		return nil
	}

	cSecretName := v.vs.GetCertSecretName(string(api.VaultClientCert))
	cSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), cSecretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			// Create new server TLS
			caSecretName := v.vs.GetCertSecretName(string(api.VaultCACert))
			caSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), caSecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// validate tls.crt
			caCertByte, exist := caSecret.Data[core.TLSCertKey]
			if !exist {
				return errors.Errorf("%s/%s does not contain tls.crt", caSecret.Namespace, caSecret.Name)
			}
			caCerts, err := cert.ParseCertsPEM(caCertByte)
			if err != nil {
				return err
			}
			if len(caCerts) == 0 {
				return errors.New("tls.crt is empty")
			}

			// validate tls.key
			caKeyByte, exist := caSecret.Data[core.TLSPrivateKeyKey]
			if !exist {
				return errors.Errorf("%s/%s does not contain tls.key", caSecret.Namespace, caSecret.Name)
			}

			caKeyInterface, err := cert.ParsePrivateKeyPEM(caKeyByte)
			if err != nil {
				return err
			}
			caKey, valid := caKeyInterface.(*rsa.PrivateKey)
			if !valid {
				klog.Infof("%v", caKey)
				return errors.New("Unsupported private key type")
			}

			cfg := cert.Config{
				CommonName:   v.vs.GetCertificateCN(api.VaultClientCert),
				Organization: []string{kubevault.GroupName},
				Usages: []x509.ExtKeyUsage{
					x509.ExtKeyUsageClientAuth,
				},
			}
			cKey, err := cert.NewPrivateKey()
			if err != nil {
				return errors.New("failed to generate key for client certificate")
			}
			cKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(cKey)
			if err != nil {
				return err
			}

			cCertificate, err := newSignedCert(cfg, cKey, caCerts[0], caKey)
			if err != nil {
				return errors.New("failed to sign client certificate")
			}
			cCertByte := cert.EncodeCertPEM(cCertificate)

			_, _, err = core_util.CreateOrPatchSecret(context.TODO(), v.kubeClient,
				metav1.ObjectMeta{
					Name:      v.vs.GetCertSecretName(string(api.VaultClientCert)),
					Namespace: v.vs.Namespace,
				},
				func(in *core.Secret) *core.Secret {
					in.Labels = v.vs.OffshootLabels()
					core_util.EnsureOwnerReference(&in.ObjectMeta, metav1.NewControllerRef(v.vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
					in.Data = map[string][]byte{
						core.TLSCertKey:       cCertByte,
						core.TLSPrivateKeyKey: cKeyByte,
						conapi.TLSCACertKey:   caCertByte,
					}
					in.Type = core.SecretTypeTLS
					return in
				}, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to create client k8s secret")
			}

		}
		return err
	}

	// validate keys
	if value, exist := cSecret.Data[conapi.TLSCACertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain ca.crt", cSecret.Namespace, cSecret.Name)
	}

	if value, exist := cSecret.Data[core.TLSCertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.crt", cSecret.Namespace, cSecret.Name)
	}

	if value, exist := cSecret.Data[core.TLSPrivateKeyKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.key", cSecret.Namespace, cSecret.Name)
	}

	return nil
}

func (v *vaultSrv) EnsureStorageTLS() error {
	if v.vs.Spec.TLS == nil {
		return nil
	}
	// If Certificate is managed by cert-manager
	if v.vs.Spec.TLS.IssuerRef != nil {
		return nil
	}
	backend, err := v.vs.Spec.Backend.GetBackendType()
	if err != nil {
		return err
	}
	if backend != api.VaultServerRaft {
		return nil
	}

	sSecretName := v.vs.GetCertSecretName(string(api.VaultStorageCert))
	sSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), sSecretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			// Create new Storage TLS
			caSecretName := v.vs.GetCertSecretName(string(api.VaultCACert))
			caSecret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), caSecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// validate tls.crt
			caCertByte, exist := caSecret.Data[core.TLSCertKey]
			if !exist {
				return errors.Errorf("%s/%s does not contain tls.crt", caSecret.Namespace, caSecret.Name)
			}
			caCerts, err := cert.ParseCertsPEM(caCertByte)
			if err != nil {
				return err
			}
			if len(caCerts) == 0 {
				return errors.New("tls.crt is empty")
			}

			// validate tls.key
			caKeyByte, exist := caSecret.Data[core.TLSPrivateKeyKey]
			if !exist {
				return errors.Errorf("%s/%s does not contain tls.key", caSecret.Namespace, caSecret.Name)
			}

			caKeyInterface, err := cert.ParsePrivateKeyPEM(caKeyByte)
			if err != nil {
				return err
			}
			caKey, valid := caKeyInterface.(*rsa.PrivateKey)
			if !valid {
				klog.Infof("%v", caKey)
				return errors.New("Unsupported private key type")
			}

			cfg := cert.Config{
				CommonName:   v.vs.GetCertificateCN(api.VaultStorageCert),
				Organization: []string{kubevault.GroupName},
				AltNames: cert.AltNames{
					DNSNames: []string{
						"localhost",
						fmt.Sprintf("*.%s.pod", v.vs.Namespace),
						fmt.Sprintf("%s.%s.svc", v.vs.ServiceName(api.VaultServerServiceVault), v.vs.Namespace),
						fmt.Sprintf("*.%s.%s.svc", v.vs.ServiceName(api.VaultServerServiceInternal), v.vs.Namespace),
						fmt.Sprintf("%s.%s.svc", v.vs.ServiceName(api.VaultServerServiceInternal), v.vs.Namespace),
					},
					IPs: []net.IP{
						net.ParseIP("127.0.0.1"),
					},
				},
				Usages: []x509.ExtKeyUsage{
					x509.ExtKeyUsageServerAuth,
					x509.ExtKeyUsageClientAuth,
				},
			}
			sKey, err := cert.NewPrivateKey()
			if err != nil {
				return errors.New("failed to generate key for storage certificate")
			}
			sKeyByte, err := cert.EncodePKCS8PrivateKeyPEM(sKey)
			if err != nil {
				return err
			}

			sCertificate, err := newSignedCert(cfg, sKey, caCerts[0], caKey)
			if err != nil {
				return errors.New("failed to sign storage certificate")
			}
			sCertByte := cert.EncodeCertPEM(sCertificate)

			_, _, err = core_util.CreateOrPatchSecret(context.TODO(), v.kubeClient,
				metav1.ObjectMeta{
					Name:      v.vs.GetCertSecretName(string(api.VaultStorageCert)),
					Namespace: v.vs.Namespace,
				},
				func(in *core.Secret) *core.Secret {
					in.Labels = v.vs.OffshootLabels()
					core_util.EnsureOwnerReference(&in.ObjectMeta, metav1.NewControllerRef(v.vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
					in.Data = map[string][]byte{
						core.TLSCertKey:       sCertByte,
						core.TLSPrivateKeyKey: sKeyByte,
						conapi.TLSCACertKey:   caCertByte,
					}
					in.Type = core.SecretTypeTLS
					return in
				}, metav1.PatchOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to create storage k8s secret")
			}

		}
		return err
	}

	// validate keys
	if value, exist := sSecret.Data[conapi.TLSCACertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain ca.crt", sSecret.Namespace, sSecret.Name)
	}

	if value, exist := sSecret.Data[core.TLSCertKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.crt", sSecret.Namespace, sSecret.Name)
	}

	if value, exist := sSecret.Data[core.TLSPrivateKeyKey]; !exist || len(value) == 0 {
		return errors.Errorf("%s/%s does not contain tls.key", sSecret.Namespace, sSecret.Name)
	}

	return nil
}

// GetConfig will return the vault config in ConfigMap
// ConfigMap will contain:
// - listener config
// - storage config
// - user provided extra config
func (v *vaultSrv) GetConfig() (*core.Secret, error) {
	configSecretName := v.vs.ConfigSecretName()
	cfgData := util.GetListenerConfig(v.vs.Spec.TLS != nil)

	storageCfg := ""
	var err error

	// TODO:
	//   - need to check while adding support for PVC
	if v.strg != nil {
		storageCfg, err = v.strg.GetStorageConfig()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get storage config")
		}
	}

	exporterCfg, err := v.exprtr.GetTelemetryConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get exporter config")
	}

	// enable ui
	uiCfg := "ui = true"

	cfgData = strings.Join([]string{cfgData, uiCfg, storageCfg, exporterCfg}, "\n")

	secret := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configSecretName,
			Namespace: v.vs.Namespace,
			Labels:    v.vs.OffshootLabels(),
		},
		StringData: map[string]string{
			filepath.Base(util.VaultConfigFile): cfgData,
		},
	}
	return secret, nil
}

// - add secret volume mount for tls secret
// - add configMap volume mount for vault config
// - add extra env, volume mount, unsealer, container, etc
func (v *vaultSrv) Apply(pt *core.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	// Add init container
	// this init container will append user provided configuration
	// file to the controller provided configuration file
	initCont := core.Container{
		Name:            util.VaultInitContainerName,
		Image:           "busybox",
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"/bin/sh"},
		Args: []string{
			"-c",
			`set -e
			cat /etc/vault/controller/vault.hcl > /etc/vault/config/vault.hcl
			echo "" >> /etc/vault/config/vault.hcl
			if [ -f /etc/vault/user/vault.hcl ]; then
				  cat /etc/vault/user/vault.hcl >> /etc/vault/config/vault.hcl
			fi`,
		},
	}

	initCont.VolumeMounts = core_util.UpsertVolumeMount(initCont.VolumeMounts,
		core.VolumeMount{
			Name:      "config",
			MountPath: filepath.Dir(util.VaultConfigFile),
		}, core.VolumeMount{
			Name:      "controller-config",
			MountPath: "/etc/vault/controller",
		})

	pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes,
		core.Volume{
			Name: "config",
			VolumeSource: core.VolumeSource{
				EmptyDir: &core.EmptyDirVolumeSource{},
			},
		}, core.Volume{
			Name: "controller-config",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: v.vs.ConfigSecretName(),
				},
			},
		})

	if v.vs.Spec.ConfigSecret != nil {
		initCont.VolumeMounts = core_util.UpsertVolumeMount(initCont.VolumeMounts, core.VolumeMount{
			Name:      "user-config",
			MountPath: "/etc/vault/user",
		})

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
			Name: "user-config",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: v.vs.Spec.ConfigSecret.Name,
				},
			},
		})
	}

	// TODO:
	// 	- feature: make tls optional if possible
	tlsSecret := v.vs.GetCertSecretName(string(api.VaultServerCert))
	pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
		Name: vaultTLSAssetVolumeName,
		VolumeSource: core.VolumeSource{
			Secret: &core.SecretVolumeSource{
				SecretName: tlsSecret,
			},
		},
	})

	var cont core.Container
	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultContainerName {
			cont = c
		}
	}

	if v.vs.Spec.TLS != nil {
		cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, core.VolumeMount{
			Name:      vaultTLSAssetVolumeName,
			MountPath: util.VaultTLSAssetDir,
		})
	}
	cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, core.VolumeMount{
		Name:      "config",
		MountPath: filepath.Dir(util.VaultConfigFile),
	})

	for indx, data := range v.vs.Spec.DataSources {
		cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, core.VolumeMount{
			Name:      "data-" + strconv.Itoa(indx),
			MountPath: "/etc/vault/data/data-" + strconv.Itoa(indx),
		})

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
			Name:         "data-" + strconv.Itoa(indx),
			VolumeSource: data,
		})
	}

	pt.Spec.InitContainers = core_util.UpsertContainer(pt.Spec.InitContainers, initCont)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)

	if v.strg != nil {
		err := v.strg.Apply(pt)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if v.unslr != nil {
		err := v.unslr.Apply(pt)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err := v.exprtr.Apply(pt, v.vs)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (v *vaultSrv) GetService() *core.Service {
	//  match with "vault" alias from the service array & place them in service object here.
	if v.vs.Spec.ServiceTemplates == nil {
		return nil
	}

	var vsTemplate api.NamedServiceTemplateSpec
	for i := range v.vs.Spec.ServiceTemplates {
		namedSpec := v.vs.Spec.ServiceTemplates[i]
		if namedSpec.Alias == api.VaultServerServiceVault {
			vsTemplate = namedSpec
		}
	}

	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.ServiceName(api.VaultServerServiceVault),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: vsTemplate.Annotations,
		},
		Spec: core.ServiceSpec{
			Selector: v.vs.OffshootSelectors(),
			Ports: []core.ServicePort{
				{
					Name:     "client",
					Protocol: core.ProtocolTCP,
					Port:     VaultClientPort,
				},
				{
					Name:     "cluster",
					Protocol: core.ProtocolTCP,
					Port:     VaultClusterPort,
				},
			},
			ClusterIP:                vsTemplate.Spec.ClusterIP,
			Type:                     vsTemplate.Spec.Type,
			ExternalIPs:              vsTemplate.Spec.ExternalIPs,
			LoadBalancerIP:           vsTemplate.Spec.LoadBalancerIP,
			LoadBalancerSourceRanges: vsTemplate.Spec.LoadBalancerSourceRanges,
			ExternalTrafficPolicy:    vsTemplate.Spec.ExternalTrafficPolicy,
			HealthCheckNodePort:      vsTemplate.Spec.HealthCheckNodePort,
			SessionAffinityConfig:    vsTemplate.Spec.SessionAffinityConfig,
		},
	}
}

func (v *vaultSrv) GetGoverningService() *core.Service {
	var inSvc api.NamedServiceTemplateSpec
	for i := range v.vs.Spec.ServiceTemplates {
		temp := v.vs.Spec.ServiceTemplates[i]
		if temp.Alias == api.VaultServerServiceInternal {
			inSvc = temp
		}
	}

	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.ServiceName(api.VaultServerServiceInternal),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: inSvc.Annotations,
		},
		Spec: core.ServiceSpec{
			Selector: v.vs.OffshootSelectors(),
			Ports: []core.ServicePort{
				{
					Name:     "client-internal",
					Protocol: core.ProtocolTCP,
					Port:     VaultClientPort,
				},
				{
					Name:     "cluster-internal",
					Protocol: core.ProtocolTCP,
					Port:     VaultClusterPort,
				},
			},
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
		},
	}
}

func (v *vaultSrv) GetStatefulSet(serviceName string, pt *core.PodTemplateSpec, vcts []core.PersistentVolumeClaim) *apps.StatefulSet {

	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: v.vs.Spec.PodTemplate.Controller.Annotations,
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    v.vs.Spec.Replicas,
			Selector:    &metav1.LabelSelector{MatchLabels: v.vs.OffshootSelectors()},
			ServiceName: serviceName,
			Template:    *pt,
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			VolumeClaimTemplates: vcts,
		},
	}
}

func (v *vaultSrv) GetServiceAccounts() []core.ServiceAccount {
	return []core.ServiceAccount{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      v.vs.ServiceAccountName(),
				Namespace: v.vs.Namespace,
				Labels:    v.vs.OffshootLabels(),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      v.vs.ServiceAccountForTokenReviewer(),
				Namespace: v.vs.Namespace,
				Labels:    v.vs.OffshootLabels(),
			},
		},
	}
}

func (v *vaultSrv) GetRBACRolesAndRoleBindings() ([]rbac.Role, []rbac.RoleBinding) {
	var roles []rbac.Role
	var rbindings []rbac.RoleBinding
	labels := v.vs.OffshootLabels()
	if v.unslr != nil {
		rList := v.unslr.GetRBAC(v.vs.Name, v.vs.Namespace)
		for _, r := range rList {
			r.Labels = core_util.UpsertMap(r.Labels, labels)
			roles = append(roles, r)

			// create corresponding role binding
			rb := rbac.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.Name,
					Namespace: r.Namespace,
					Labels:    r.Labels,
				},
				RoleRef: rbac.RoleRef{
					APIGroup: rbac.GroupName,
					Kind:     "Role",
					Name:     r.Name,
				},
				Subjects: []rbac.Subject{
					{
						Kind:      rbac.ServiceAccountKind,
						Name:      v.vs.ServiceAccountName(),
						Namespace: v.vs.Namespace,
					},
				},
			}
			rbindings = append(rbindings, rb)
		}
	}
	return roles, rbindings
}

func (v *vaultSrv) GetRBACClusterRoleBinding() rbac.ClusterRoleBinding {
	return rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   meta_util.NameWithPrefix(v.vs.Namespace+"-"+v.vs.Name, "k8s-token-reviewer"),
			Labels: v.vs.OffshootLabels(),
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     "system:auth-delegator",
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      v.vs.ServiceAccountForTokenReviewer(),
				Namespace: v.vs.Namespace,
			},
		},
	}
}

func (v *vaultSrv) GetPodTemplate(c core.Container, saName string) *core.PodTemplateSpec {
	// If SecurityContext for PodTemplate is not specified, SecurityContext.FSGroup is set to default value 65535
	if v.vs.Spec.PodTemplate.Spec.SecurityContext == nil {
		v.vs.Spec.PodTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	v.vs.Spec.PodTemplate.Spec.SecurityContext.FSGroup = &defaultFsGroup

	return &core.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.Name,
			Labels:      v.vs.OffshootSelectors(),
			Namespace:   v.vs.Namespace,
			Annotations: v.vs.Spec.PodTemplate.Annotations,
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				c,
			},
			ServiceAccountName: saName,
			NodeSelector:       v.vs.Spec.PodTemplate.Spec.NodeSelector,
			Affinity:           v.vs.Spec.PodTemplate.Spec.Affinity,
			SchedulerName:      v.vs.Spec.PodTemplate.Spec.SchedulerName,
			Tolerations:        v.vs.Spec.PodTemplate.Spec.Tolerations,
			ImagePullSecrets:   v.vs.Spec.PodTemplate.Spec.ImagePullSecrets,
			PriorityClassName:  v.vs.Spec.PodTemplate.Spec.PriorityClassName,
			Priority:           v.vs.Spec.PodTemplate.Spec.Priority,
			SecurityContext:    v.vs.Spec.PodTemplate.Spec.SecurityContext,
		},
	}
}

func (v *vaultSrv) GetContainer() core.Container {
	container := core.Container{
		Name:            util.VaultContainerName,
		Image:           v.config.Image,
		ImagePullPolicy: v.config.ImagePullPolicy,
		Command: []string{
			"/bin/vault",
			"server",
			"-config=" + util.VaultConfigFile, "-log-level=debug",
		},
		Env: []core.EnvVar{
			{
				Name: "HOSTNAME",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
			},
			{
				Name: "HOST_IP",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.hostIP",
					},
				},
			},
			{
				Name: "POD_IP",
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
			{
				Name:  EnvVaultAPIAddr,
				Value: util.VaultServiceURL(v.vs.Name, v.vs.Namespace, VaultClientPort),
			},
			{
				Name:  EnvVaultClusterAddr,
				Value: util.VaultServiceURL(v.vs.Name, v.vs.Namespace, VaultClusterPort),
			},
		},
		SecurityContext: &core.SecurityContext{
			Capabilities: &core.Capabilities{
				// Vault requires mlock syscall to work.
				// Without this it would fail "Error initializing core: Failed to lock memory: cannot allocate memory"
				Add: []core.Capability{"IPC_LOCK"},
			},
		},
		Ports: []core.ContainerPort{{
			Name:          "vault-port",
			ContainerPort: int32(VaultClientPort),
		}, {
			Name:          "cluster-port",
			ContainerPort: int32(VaultClusterPort),
		}},
		ReadinessProbe: func() *core.Probe {
			if v.vs.Spec.PodTemplate.Spec.ReadinessProbe != nil {
				return v.vs.Spec.PodTemplate.Spec.ReadinessProbe
			}
			return &core.Probe{
				Handler: core.Handler{
					HTTPGet: &core.HTTPGetAction{
						Path: "/v1/sys/health?standbyok=true&perfstandbyok=true",
						Port: intstr.FromInt(VaultClientPort),
						Scheme: func() core.URIScheme {
							if v.vs.Spec.TLS != nil {
								return core.URISchemeHTTPS
							}
							return core.URISchemeHTTP
						}(),
					},
				},
				InitialDelaySeconds: 10,
				TimeoutSeconds:      10,
				PeriodSeconds:       10,
				FailureThreshold:    5,
			}
		}(),
		Resources: v.vs.Spec.PodTemplate.Spec.Resources,
	}

	if v.vs.Spec.TLS != nil {
		container.Env = core_util.UpsertEnvVars(container.Env,
			core.EnvVar{
				Name:  EnvVaultCACert,
				Value: fmt.Sprintf("%s%s", util.VaultTLSAssetDir, conapi.TLSCACertKey),
			})
	}

	return container
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func newSignedCert(cfg cert.Config, key *rsa.PrivateKey, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(certlib.Duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
		ExtraExtensions: []pkix.Extension{
			{
				Id: oidExtensionSubjectAltName,
			},
		},
	}
	certTmpl.ExtraExtensions[0].Value, err = marshalSANs(cfg.AltNames.DNSNames, nil, cfg.AltNames.IPs)
	if err != nil {
		return nil, err
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certDERBytes)
}

var (
	oidExtensionSubjectAltName = []int{2, 5, 29, 17}
)

// marshalSANs marshals a list of addresses into a the contents of an X.509
// SubjectAlternativeName extension.
func marshalSANs(dnsNames, emailAddresses []string, ipAddresses []net.IP) (derBytes []byte, err error) {
	var rawValues []asn1.RawValue
	for _, name := range dnsNames {
		rawValues = append(rawValues, asn1.RawValue{Tag: 2, Class: 2, Bytes: []byte(name)})
	}
	for _, email := range emailAddresses {
		rawValues = append(rawValues, asn1.RawValue{Tag: 1, Class: 2, Bytes: []byte(email)})
	}
	for _, rawIP := range ipAddresses {
		// If possible, we always want to encode IPv4 addresses in 4 bytes.
		ip := rawIP.To4()
		if ip == nil {
			ip = rawIP
		}
		rawValues = append(rawValues, asn1.RawValue{Tag: 7, Class: 2, Bytes: ip})
	}
	// https://github.com/floragunncom/search-guard-docs/blob/master/tls_certificates_production.md#using-an-oid-value-as-san-entry
	// https://github.com/floragunncom/search-guard-ssl/blob/a2d1e8e9b25a10ecaf1cb47e48cf04328af7d7fb/example-pki-scripts/gen_node_cert.sh#L55
	// Adds AltName: OID: 1.2.3.4.5.5
	// ref: https://stackoverflow.com/a/47917273/244009
	rawValues = append(rawValues, asn1.RawValue{FullBytes: []byte{0x88, 0x05, 0x2A, 0x03, 0x04, 0x05, 0x05}})
	return asn1.Marshal(rawValues)
}
