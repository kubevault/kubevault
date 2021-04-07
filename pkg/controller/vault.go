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
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	conapi "kubevault.dev/apimachinery/apis"
	capi "kubevault.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	cs "kubevault.dev/apimachinery/client/clientset/versioned"
	"kubevault.dev/operator/pkg/vault/exporter"
	"kubevault.dev/operator/pkg/vault/storage"
	"kubevault.dev/operator/pkg/vault/unsealer"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/pkg/errors"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
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
	GetServerTLS() (*core.Secret, []byte, error)
	GetConfig() (*core.Secret, error)
	Apply(pt *core.PodTemplateSpec) error
	GetService() *core.Service
	GetHeadlessService() *core.Service
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
	stfStrg    storage.StatefulStorage
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
	var stfStrg storage.StatefulStorage
	strg, err := storage.NewStorage(kc, vs)
	if err != nil {
		var er error
		stfStrg, er = storage.NewStatefulStorage(kc, vs)
		if er != nil {
			return nil, err
		}
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
		stfStrg:    stfStrg,
		unslr:      unslr,
		exprtr:     exprtr,
		kubeClient: kc,
		config:     version.Spec.Vault,
	}, nil
}

// GetServerTLS will return a secret containing vault server tls assets
// secret contains following data:
// 	- ca.crt : <ca.crt-used-to-sign-vault-server-cert>
//  - server.crt -> tls.crt : <vault-server-cert>
//  - server.key -> tls.key : <vault-server-key>
//
// if user provide TLS secrets, then it will be used.
// Otherwise self signed certificates will be used
func (v *vaultSrv) GetServerTLS() (*core.Secret, []byte, error) {
	// Goal -> 3 Cases:
	//  - User can provide custom TLS secret - do this later
	//		- Validate & use it
	// 	- User can only provide Secret Name - do this later
	//		- Use the secret name to create secret
	//  - User can provide nothing - do this now
	//		- Create Secret with the default name
	// tls:
	// - alias: vault
	// 	 secretName: <name>
	// 	 ... ... ..

	//tls := v.vs.Spec.TLS
	//if tls != nil && tls.Certificates != nil {
	//	secretName := v.vs.GetCertSecretName("vault")
	//	secret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	//
	//	byt, ok := secret.Data[conapi.TLSCACertKey]
	//	if !ok {
	//		return nil, nil, errors.New("missing ca.crt in vault secret")
	//	}
	//	return secret, byt, err
	//}
	//
	//if v.vs.Spec.TLS == nil {
	//	v.vs.Spec.TLS = &kmapi.TLSConfig{
	//		IssuerRef: tls.IssuerRef,
	//		Certificates: tls.Certificates,
	//	}
	//}

	//tlsSecretName := v.vs.Spec.TLS.TLSSecret
	//
	//sr, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), tlsSecretName, metav1.GetOptions{})
	//if err == nil {
	//	glog.Infof("secret %s/%s already exists", v.vs.Namespace, tlsSecretName)
	//	return sr, v.vs.Spec.TLS.CABundle, nil
	//}

	// get the secretName
	secretName := v.vs.GetCertSecretName("vault")
	secret, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil && !errors2.IsNotFound(err) {
		return nil, nil, err
	} else if errors2.IsNotFound(err) {
		// create secret
		store, err := certstore.New(blobfs.NewInMemoryFS(), filepath.Join("", "pki"))
		if err != nil {
			return nil, nil, errors.Wrap(err, "certificate store create error")
		}

		err = store.NewCA()
		if err != nil {
			return nil, nil, errors.Wrap(err, "ca certificate create error")
		}

		// ref: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/

		altNames := cert.AltNames{
			DNSNames: []string{
				"localhost",
				fmt.Sprintf("*.%s.pod", v.vs.Namespace),
				fmt.Sprintf("%s.%s.svc", v.vs.Name, v.vs.Namespace),
			},
			IPs: []net.IP{
				net.ParseIP("127.0.0.1"),
			},
		}

		// XXX when required only

		altNames.DNSNames = append(
			altNames.DNSNames,
			"*.vault-internal",
			fmt.Sprintf("%s.vault-internal.%s.svc", v.vs.Name, v.vs.Namespace),
		)

		// XXX allow both kind of certificates to be made depending on the usage.
		// srvCrt, srvKey, err := store.NewServerCertPairBytes(altNames)
		srvCrt, srvKey, err := store.NewPeerCertPairBytes(altNames)
		if err != nil {
			return nil, nil, errors.Wrap(err, "vault server create crt/key pair error")
		}

		tlsSecret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: v.vs.Namespace,
				Labels:    v.vs.OffshootLabels(),
			},
			Data: map[string][]byte{
				core.TLSCertKey:       srvCrt, // tls.crt
				core.TLSPrivateKeyKey: srvKey, // tls.key
				conapi.TLSCACertKey:   store.CACertBytes(),
			},
			Type: core.SecretTypeTLS,
		}
		// create the secret
		secret, err = v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Create(context.TODO(), tlsSecret, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, errors.Wrap(err, "secret creating error")
		}

		return secret, store.CACertBytes(), nil
	}

	// Already secret exist
	// validate it:
	// - check tls.crt, tls.key, ca.crt exist
	data := secret.Data
	if _, ok := data[core.TLSCertKey]; !ok {
		return nil, nil, errors.New("tls.crt is missing")
	}
	if _, ok := data[core.TLSPrivateKeyKey]; !ok {
		return nil, nil, errors.New("tls.key is missing")
	}
	ca, ok := data[conapi.TLSCACertKey]
	if !ok {
		return nil, nil, errors.New("ca.crt is missing")
	}
	return secret, ca, nil
}

// GetConfig will return the vault config in ConfigMap
// ConfigMap will contain:
// - listener config
// - storage config
// - user provided extra config
func (v *vaultSrv) GetConfig() (*core.Secret, error) {
	configSecretName := v.vs.ConfigMapName()
	cfgData := util.GetListenerConfig()

	storageCfg := ""
	var err error

	// TODO:
	//   - need to check while adding support for PVC
	if v.strg != nil {
		storageCfg, err = v.strg.GetStorageConfig()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get storage config")
		}
	} else {
		storageCfg, err = v.stfStrg.GetStorageConfig()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get storage config")
		}
	}

	exporterCfg, err := v.exprtr.GetTelemetryConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get exporter config")
	}

	cfgData = strings.Join([]string{cfgData, storageCfg, exporterCfg}, "\n")

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
					SecretName: v.vs.ConfigMapName(),
				},
			},
		})

	if v.vs.Spec.TLS != nil {
		initCont.VolumeMounts = core_util.UpsertVolumeMount(initCont.VolumeMounts, core.VolumeMount{
			Name:      "user-config",
			MountPath: "/etc/vault/user",
		})

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
			Name: "user-config",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{},
			},
		})
	}

	// TODO:
	// 	- feature: make tls optional if possible
	tlsSecret := v.vs.GetCertSecretName("vault")
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

	cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, core.VolumeMount{
		Name:      vaultTLSAssetVolumeName,
		MountPath: util.VaultTLSAssetDir,
	}, core.VolumeMount{
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
	} else {
		err := v.stfStrg.Apply(pt)
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
		if namedSpec.Alias == "vault" {
			vsTemplate = namedSpec
		}
	}

	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
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

func (v *vaultSrv) GetHeadlessService() *core.Service {
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

func (v *vaultSrv) GetDeployment(pt *core.PodTemplateSpec) *apps.Deployment {
	if v.strg == nil {
		return nil
	}

	return &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: v.vs.Spec.PodTemplate.Controller.Annotations,
		},
		Spec: apps.DeploymentSpec{
			Replicas: v.vs.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: v.vs.OffshootSelectors()},
			Template: *pt,
			Strategy: apps.DeploymentStrategy{
				Type: apps.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &apps.RollingUpdateDeployment{
					MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				},
			},
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
	return core.Container{
		Name:            util.VaultContainerName,
		Image:           v.config.Image,
		ImagePullPolicy: v.config.ImagePullPolicy,
		Command: []string{
			"/bin/vault",
			"server",
			"-config=" + util.VaultConfigFile,
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
			{
				Name:  EnvVaultCACert,
				Value: fmt.Sprintf("%s%s", util.VaultTLSAssetDir, conapi.TLSCACertKey),
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
		ReadinessProbe: &core.Probe{
			Handler: core.Handler{
				HTTPGet: &core.HTTPGetAction{
					Path:   "/v1/sys/health?standbyok=true&perfstandbyok=true",
					Port:   intstr.FromInt(VaultClientPort),
					Scheme: core.URISchemeHTTPS,
				},
			},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			PeriodSeconds:       10,
			FailureThreshold:    5,
		},
		Resources: v.vs.Spec.PodTemplate.Spec.Resources,
	}
}
