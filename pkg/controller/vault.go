package controller

import (
	"fmt"
	"path/filepath"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/golang/glog"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	cs "github.com/kubevault/operator/client/clientset/versioned"
	"github.com/kubevault/operator/pkg/vault/exporter"
	"github.com/kubevault/operator/pkg/vault/storage"
	"github.com/kubevault/operator/pkg/vault/unsealer"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
)

const (
	EnvVaultAddr            = "VAULT_API_ADDR"
	EnvVaultClusterAddr     = "VAULT_CLUSTER_ADDR"
	VaultClientPort         = 8200
	VaultClusterPort        = 8201
	vaultTLSAssetVolumeName = "vault-tls-secret"
	CaCertName              = "ca.crt"
	ServerCertName          = "server.crt"
	ServerkeyName           = "server.key"
)

type Vault interface {
	GetServerTLS() (*core.Secret, error)
	GetConfig() (*core.ConfigMap, error)
	Apply(pt *core.PodTemplateSpec) error
	GetService() *core.Service
	GetDeployment(pt *core.PodTemplateSpec) *apps.Deployment
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
	image      string
}

func NewVault(vs *api.VaultServer, config *rest.Config, kc kubernetes.Interface, vc cs.Interface) (Vault, error) {
	version, err := vc.CatalogV1alpha1().VaultServerVersions().Get(string(vs.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get vault server version")
	}

	// it is required to have storage
	strg, err := storage.NewStorage(kc, vs)
	if err != nil {
		return nil, err
	}

	// it is not required to have unsealer
	unslr, err := unsealer.NewUnsealerService(config, vs, version.Spec.Unsealer.Image)
	if err != nil {
		return nil, err
	}

	exprtr, err := exporter.NewExporter(version.Spec.Exporter.Image)
	if err != nil {
		return nil, err
	}
	return &vaultSrv{
		vs:         vs,
		strg:       strg,
		unslr:      unslr,
		exprtr:     exprtr,
		kubeClient: kc,
		image:      version.Spec.Vault.Image,
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
func (v *vaultSrv) GetServerTLS() (*core.Secret, error) {
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

	tlsSr := &core.Secret{
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
func (v *vaultSrv) GetConfig() (*core.ConfigMap, error) {
	configMapName := v.vs.ConfigMapName()
	cfgData := util.GetListenerConfig()

	storageCfg, err := v.strg.GetStorageConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get storage config")
	}

	exporterCfg, err := v.exprtr.GetTelemetryConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get exporter config")
	}

	cfgData = fmt.Sprintf("%s\n%s\n%s", cfgData, storageCfg, exporterCfg)

	configM := &core.ConfigMap{
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
				ConfigMap: &core.ConfigMapVolumeSource{
					LocalObjectReference: core.LocalObjectReference{
						Name: v.vs.ConfigMapName(),
					},
				},
			},
		})

	if v.vs.Spec.ConfigSource != nil {
		initCont.VolumeMounts = core_util.UpsertVolumeMount(initCont.VolumeMounts, core.VolumeMount{
			Name:      "user-config",
			MountPath: "/etc/vault/user",
		})

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
			Name:         "user-config",
			VolumeSource: *v.vs.Spec.ConfigSource,
		})
	}

	tlsSecret := v.vs.TLSSecretName()
	if v.vs.Spec.TLS != nil && v.vs.Spec.TLS.TLSSecret != "" {
		tlsSecret = v.vs.Spec.TLS.TLSSecret
	}

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

	err = v.exprtr.Apply(pt, v.vs.Spec.Monitor)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (v *vaultSrv) GetService() *core.Service {
	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: v.vs.Spec.ServiceTemplate.Annotations,
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
				{
					Name:       exporter.PrometheusExporterPortName,
					Protocol:   core.ProtocolTCP,
					Port:       exporter.VaultExporterFetchMetricsPort,
					TargetPort: intstr.FromInt(exporter.VaultExporterFetchMetricsPort),
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

func (v *vaultSrv) GetDeployment(pt *core.PodTemplateSpec) *apps.Deployment {
	return &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        v.vs.OffshootName(),
			Namespace:   v.vs.Namespace,
			Labels:      v.vs.OffshootLabels(),
			Annotations: v.vs.Spec.PodTemplate.Controller.Annotations,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &v.vs.Spec.Nodes,
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
			Name:      v.vs.Name + "-k8s-token-reviewer",
			Namespace: v.vs.Namespace,
			Labels:    v.vs.OffshootLabels(),
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
		Name:  util.VaultContainerName,
		Image: v.image,
		Command: []string{
			"/bin/vault",
			"server",
			"-config=" + util.VaultConfigFile,
		},
		Env: []core.EnvVar{
			{
				Name:  EnvVaultAddr,
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
		ReadinessProbe: &core.Probe{
			Handler: core.Handler{
				HTTPGet: &core.HTTPGetAction{
					Path:   "/v1/sys/health",
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
