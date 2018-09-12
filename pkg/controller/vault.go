package controller

import (
	"fmt"
	"path/filepath"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/golang/glog"
	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/storage"
	"github.com/kubevault/operator/pkg/vault/unsealer"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/cert"
)

const (
	EnvVaultAddr            = "VAULT_API_ADDR"
	EnvVaultClusterAddr     = "VAULT_CLUSTER_ADDR"
	VaultPort               = 8200
	VaultClusterPort        = 8201
	vaultTLSAssetVolumeName = "vault-tls-secret"
	CaCertName              = "ca.crt"
	ServerCertName          = "server.crt"
	ServerkeyName           = "server.key"
)

type Vault interface {
	GetServerTLS() (*corev1.Secret, error)
	GetConfig() (*corev1.ConfigMap, error)
	Apply(pt *corev1.PodTemplateSpec) error
	GetService() *corev1.Service
	GetDeployment(pt *corev1.PodTemplateSpec) *appsv1.Deployment
	GetServiceAccount() *corev1.ServiceAccount
	GetRBACRoles() []rbacv1.Role
	GetPodTemplate(c corev1.Container, saName string) *corev1.PodTemplateSpec
	GetContainer() corev1.Container
}

type vaultSrv struct {
	vs         *api.VaultServer
	strg       storage.Storage
	unslr      unsealer.Unsealer
	kubeClient kubernetes.Interface
}

func NewVault(vs *api.VaultServer, kc kubernetes.Interface) (Vault, error) {
	// it is required to have storage
	strg, err := storage.NewStorage(kc, vs)
	if err != nil {
		return nil, err
	}

	// it is not required to have unsealer
	unslr, err := unsealer.NewUnsealerService(vs.Spec.Unsealer)
	if err != nil {
		return nil, err
	}

	return &vaultSrv{
		vs:         vs,
		strg:       strg,
		unslr:      unslr,
		kubeClient: kc,
	}, nil
}

// GetServerTLS will return a secret containing vault server tls assets
// secret contains following data:
// 	- ca.crt : <ca.crt-used-to-sign-vault-server-cert>
//  - server.crt : <vault-server-cert>
//  - server.key : <vault-server-key>
//
// if user provide TLS secrets, then it will be used.
// Otherwise self signed certificates will be used
func (v *vaultSrv) GetServerTLS() (*corev1.Secret, error) {
	tls := v.vs.Spec.TLS
	if tls != nil && tls.TLSSecret != "" {
		sr, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(tls.TLSSecret, metav1.GetOptions{})
		return sr, err
	}

	tlsSecretName := v.vs.TLSSecretName()
	sr, err := v.kubeClient.CoreV1().Secrets(v.vs.Namespace).Get(tlsSecretName, metav1.GetOptions{})
	if err == nil {
		glog.Infof("secret %s/%s already exists", v.vs.Namespace, tlsSecretName)
		return sr, nil
	}

	store, err := certstore.NewCertStore(afero.NewMemMapFs(), filepath.Join("", "pki"))
	if err != nil {
		return nil, errors.Wrap(err, "certificate store create error")
	}

	err = store.NewCA()
	if err != nil {
		return nil, errors.Wrap(err, "ca certificate create error")
	}

	// ref: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
	altNames := cert.AltNames{
		DNSNames: []string{
			"localhost",
			fmt.Sprintf("*.%s.pod", v.vs.Namespace),
			fmt.Sprintf("%s.%s.svc", v.vs.Name, v.vs.Namespace),
		},
	}

	srvCrt, srvKey, err := store.NewServerCertPairBytes(altNames)
	if err != nil {
		return nil, errors.Wrap(err, "vault server create crt/key pair error")
	}

	tlsSr := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tlsSecretName,
			Namespace: v.vs.Namespace,
			Labels:    v.vs.OffshootLabels(),
		},
		Data: map[string][]byte{
			CaCertName:     store.CACertBytes(),
			ServerCertName: srvCrt,
			ServerkeyName:  srvKey,
		},
	}
	return tlsSr, nil
}

// GetConfig will return the vault config in ConfigMap
// ConfigMap will contain:
// - listener config
// - storage config
// - user provided extra config
func (v *vaultSrv) GetConfig() (*corev1.ConfigMap, error) {
	configMapName := v.vs.ConfigMapName()
	cfgData := util.GetListenerConfig()

	storageCfg, err := v.strg.GetStorageConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get storage config")
	}
	cfgData = fmt.Sprintf("%s\n%s", cfgData, storageCfg)

	configM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: v.vs.Namespace,
			Labels:    v.vs.OffshootLabels(),
		},
		Data: map[string]string{
			filepath.Base(util.VaultConfigFile): cfgData,
		},
	}
	return configM, nil
}

// - add secret volume mount for tls secret
// - add configMap volume mount for vault config
// - add extra env, volume mount, unsealer contianer etc
func (v *vaultSrv) Apply(pt *corev1.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	// Add init container
	// this init container will append user provided configuration
	// file to the controller provided configuration file
	initCont := corev1.Container{
		Name:    util.VaultInitContainerImageName(),
		Image:   "busybox",
		Command: []string{"/bin/sh"},
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
		corev1.VolumeMount{
			Name:      "config",
			MountPath: filepath.Dir(util.VaultConfigFile),
		}, corev1.VolumeMount{
			Name:      "controller-config",
			MountPath: "/etc/vault/controller",
		})

	pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes,
		corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}, corev1.Volume{
			Name: "controller-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: v.vs.ConfigMapName(),
					},
				},
			},
		})

	if v.vs.Spec.ConfigSource != nil {
		initCont.VolumeMounts = core_util.UpsertVolumeMount(initCont.VolumeMounts, corev1.VolumeMount{
			Name:      "user-config",
			MountPath: "/etc/vault/user",
		})

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, corev1.Volume{
			Name:         "user-config",
			VolumeSource: *v.vs.Spec.ConfigSource,
		})
	}

	tlsSecret := v.vs.TLSSecretName()
	if v.vs.Spec.TLS != nil && v.vs.Spec.TLS.TLSSecret != "" {
		tlsSecret = v.vs.Spec.TLS.TLSSecret
	}

	pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, corev1.Volume{
		Name: vaultTLSAssetVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: tlsSecret,
			},
		},
	})

	var cont corev1.Container
	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultImageName() {
			cont = c
		}
	}

	cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, corev1.VolumeMount{
		Name:      vaultTLSAssetVolumeName,
		MountPath: util.VaultTLSAssetDir,
	}, corev1.VolumeMount{
		Name:      "config",
		MountPath: filepath.Dir(util.VaultConfigFile),
	})

	pt.Spec.InitContainers = core_util.UpsertContainer(pt.Spec.InitContainers, initCont)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)

	err := v.strg.Apply(pt)
	if err != nil {
		return errors.WithStack(err)
	}

	if v.unslr != nil {
		err = v.unslr.Apply(pt)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (v *vaultSrv) GetService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: v.vs.Spec.ServiceTemplate.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: v.vs.OffshootSelectors(),
			Ports: []corev1.ServicePort{
				{
					Name:     "vault-port",
					Protocol: corev1.ProtocolTCP,
					Port:     VaultPort,
				},
				{
					Name:     "cluster-port",
					Protocol: corev1.ProtocolTCP,
					Port:     VaultClusterPort,
				},
			},
			ClusterIP:                v.vs.Spec.ServiceTemplate.Spec.ClusterIP,
			Type:                     v.vs.Spec.ServiceTemplate.Spec.Type,
			ExternalIPs:              v.vs.Spec.ServiceTemplate.Spec.ExternalIPs,
			LoadBalancerIP:           v.vs.Spec.ServiceTemplate.Spec.LoadBalancerIP,
			LoadBalancerSourceRanges: v.vs.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges,
			ExternalTrafficPolicy:    v.vs.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy,
			HealthCheckNodePort:      v.vs.Spec.ServiceTemplate.Spec.HealthCheckNodePort,
		},
	}
}

func (v *vaultSrv) GetDeployment(pt *corev1.PodTemplateSpec) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: v.vs.Spec.PodTemplate.Controller.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &v.vs.Spec.Nodes,
			Selector: &metav1.LabelSelector{MatchLabels: v.vs.OffshootSelectors()},
			Template: *pt,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				},
			},
		},
	}
}

func (v *vaultSrv) GetServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.vs.OffshootName(),
			Namespace: v.vs.Namespace,
			Labels:    v.vs.OffshootLabels(),
		},
	}
}

func (v *vaultSrv) GetRBACRoles() []rbacv1.Role {
	var roles []rbacv1.Role
	labels := v.vs.OffshootLabels()
	if v.unslr != nil {
		rList := v.unslr.GetRBAC(v.vs.Namespace)
		for _, r := range rList {
			r.Labels = core_util.UpsertMap(r.Labels, labels)
			roles = append(roles, r)
		}
	}
	return roles
}

func (v *vaultSrv) GetPodTemplate(c corev1.Container, saName string) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.Name,
			Labels:      v.vs.OffshootSelectors(),
			Namespace:   v.vs.Namespace,
			Annotations: v.vs.Spec.PodTemplate.Annotations,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
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

func (v *vaultSrv) GetContainer() corev1.Container {
	return corev1.Container{
		Name:  util.VaultImageName(),
		Image: util.VaultImage(v.vs),
		Command: []string{
			"/bin/vault",
			"server",
			"-config=" + util.VaultConfigFile,
		},
		Env: []corev1.EnvVar{
			{
				Name:  EnvVaultAddr,
				Value: util.VaultServiceURL(v.vs.Name, v.vs.Namespace, VaultPort),
			},
			{
				Name:  EnvVaultClusterAddr,
				Value: util.VaultServiceURL(v.vs.Name, v.vs.Namespace, VaultClusterPort),
			},
		},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				// Vault requires mlock syscall to work.
				// Without this it would fail "Error initializing core: Failed to lock memory: cannot allocate memory"
				Add: []corev1.Capability{"IPC_LOCK"},
			},
		},
		Ports: []corev1.ContainerPort{{
			Name:          "vault-port",
			ContainerPort: int32(VaultPort),
		}, {
			Name:          "cluster-port",
			ContainerPort: int32(VaultClusterPort),
		}},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/v1/sys/health",
					Port:   intstr.FromInt(VaultPort),
					Scheme: corev1.URISchemeHTTPS,
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
