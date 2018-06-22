package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/appscode/kubernetes-webhook-util/admission"
	hooks "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	webhook "github.com/appscode/kubernetes-webhook-util/admission/v1beta1/generic"
	kutilappsv1beta1 "github.com/appscode/kutil/apps/v1beta1"
	"github.com/appscode/kutil/tools/certstore"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	"github.com/kube-vault/operator/apis/core"
	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	"github.com/kube-vault/operator/client/clientset/versioned/scheme"
	patchutil "github.com/kube-vault/operator/client/clientset/versioned/typed/core/v1alpha1/util"
	"github.com/kube-vault/operator/pkg/vault/storage"
	"github.com/kube-vault/operator/pkg/vault/unsealer"
	"github.com/kube-vault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/reference"
	"k8s.io/client-go/util/cert"
)

const (
	EnvVaultAddr          = "VAULT_API_ADDR"
	EnvVaultClusterAddr   = "VAULT_CLUSTER_ADDR"
	VaultPort             = 8200
	VaultClusterPort      = 8201
	VaultConfigVolumeName = "vault-config"
	VaultTlsSecretName    = "vault-tls-secret"
	CaCertName            = "ca.crt"
	ServerCertName        = "server.crt"
	ServerkeyName         = "server.key"
)

func (c *VaultController) NewVaultServerWebhook() hooks.AdmissionHook {
	return webhook.NewGenericWebhook(
		schema.GroupVersionResource{
			Group:    "admission.core.kube-vault.com",
			Version:  "v1alpha1",
			Resource: "vaultservers",
		},
		"vaultserver",
		[]string{core.GroupName},
		api.SchemeGroupVersion.WithKind("VaultServer"),
		nil,
		&admission.ResourceHandlerFuncs{
			CreateFunc: func(obj runtime.Object) (runtime.Object, error) {
				return nil, obj.(*api.VaultServer).IsValid()
			},
			UpdateFunc: func(oldObj, newObj runtime.Object) (runtime.Object, error) {
				return nil, newObj.(*api.VaultServer).IsValid()
			},
		},
	)
}

func (c *VaultController) initVaultServerWatcher() {
	c.vsInformer = c.extInformerFactory.Core().V1alpha1().VaultServers().Informer()
	c.vsQueue = queue.New("VaultServer", c.MaxNumRequeues, c.NumThreads, c.runVaultServerInjector)

	// TODO: Add a custom event handler
	c.vsInformer.AddEventHandler(queue.DefaultEventHandler(c.vsQueue.GetQueue()))
	c.vsLister = c.extInformerFactory.Core().V1alpha1().VaultServers().Lister()
}

// runVaultSeverInjector gets the vault server object indexed by the key from cache
// and initializes, reconciles or garbage collects the vault cluster as needed.
func (c *VaultController) runVaultServerInjector(key string) error {
	obj, exists, err := c.vsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a VaultServer, so that we will see a delete for one d
		glog.Warningf("VaultServer %s does not exist anymore\n", key)

		_, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}

		// stop vault status monitor
		if cancel, ok := c.ctxCancels[name]; ok {
			cancel()
			delete(c.ctxCancels, name)
		}

	} else {
		vault := obj.(*api.VaultServer)

		glog.Infof("Sync/Add/Update for VaultServer %s/%s\n", vault.GetNamespace(), vault.GetName())
		// glog.Infoln(vault.Name, vault.Namespace)

		// TODO : initializer or validation/mutating webhook
		// will be deprecated
		changed := vault.SetDefaults()
		if changed {
			_, _, err = patchutil.CreateOrPatchVaultServer(c.extClient.CoreV1alpha1(), vault.ObjectMeta, func(v *api.VaultServer) *api.VaultServer {
				v.SetDefaults()
				return v
			})
			if err != nil {
				return errors.Wrap(err, "unable to patch vaultServer")
			}
		}

		err := c.reconcileVault(vault)
		if err != nil {
			return err
		}
	}
	return nil
}

// reconcileVault reconciles the vault cluster's state to the spec specified by v
// by preparing the TLS secrets, deploying vault cluster,
// and finally updating the vault deployment if needed.
func (c *VaultController) reconcileVault(v *api.VaultServer) error {
	d, err := c.kubeClient.AppsV1beta1().Deployments(v.Namespace).Get(v.Name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		//deploy vault

		err = c.prepareVaultTLSSecrets(v)
		if err != nil {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeWarning,
					"prepare vault tls failed",
					err.Error(),
				)
			}
			return errors.Wrap(err, "prepare vault tls secret error")
		} else {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeNormal,
					"vault tls secret created",
					fmt.Sprintf("vault tls secret '%s' created successfully", VaultTlsSecretName),
				)
			}
		}

		err = c.prepareConfig(v)
		if err != nil {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeWarning,
					"vault configuration failed",
					err.Error(),
				)
			}
			return errors.Wrap(err, "prepare vault config error")

		} else {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeNormal,
					"vault configuration created",
					fmt.Sprintf("configMap '%s' for vault configuration created successfully", util.ConfigMapNameForVault(v)),
				)
			}
		}

		err = c.DeployVault(v)
		if err != nil {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeWarning,
					"vault deploy failed",
					err.Error(),
				)
			}
			return errors.Wrap(err, "vault deploy error")

		} else {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeNormal,
					"deployment created successfully",
					fmt.Sprintf("deployment '%s' for vaultServer created successfully", v.GetName()),
				)
			}
		}

	} else if err == nil {
		// if deployment is created for vaultserver, then sync specifications
		// else give an error

		// use image to determine whether this deployment is for vaultserver
		if util.RemoveImageTag(d.Spec.Template.Spec.Containers[0].Image) != v.Spec.BaseImage {
			fmt.Println(util.RemoveImageTag(d.Spec.Template.Spec.Containers[0].Name), v.Spec.BaseImage)
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeWarning,
					"deployment exists",
					fmt.Sprintf("deployment with same name of vaultserver '%s' already exists", v.GetName()),
				)
			}

			return errors.Errorf("deployment with same name of vaultserver '%s' already exists", v.GetName())
		}

		if *d.Spec.Replicas != v.Spec.Nodes {
			d.Spec.Replicas = &(v.Spec.Nodes)
			_, err = c.kubeClient.AppsV1beta1().Deployments(v.Namespace).Update(d)
			if err != nil {
				if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
					c.recorder.Eventf(
						ref,
						corev1.EventTypeWarning,
						"deployment update failed",
						err.Error(),
					)
				}

				return errors.Wrapf(err, "failed to update size of deployment '%s'", d.Name)
			}
		}

		err = c.syncUpgrade(v, d)
		if err != nil {
			if ref, err2 := reference.GetReference(scheme.Scheme, v); err2 == nil {
				c.recorder.Eventf(
					ref,
					corev1.EventTypeWarning,
					"sync between vaultServer and deployments failed",
					err.Error(),
				)
			}
			return errors.Wrap(err, "sync vaultServer and deployments error")
		}
	} else {
		return errors.Wrap(err, "get deployments error")
	}

	if _, ok := c.ctxCancels[v.Name]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		c.ctxCancels[v.Name] = cancel
		go c.monitorAndUpdateStatus(ctx, v)
	}

	return nil
}

// DeployVault deploys a vault server.
// DeployVault is a multi-steps process. It creates the deployment, the service, service account and
// other related Kubernetes objects for Vault. Any intermediate step can fail.
//
// DeployVault is idempotent. If an object already exists, this function will ignore creating
// it and return no error. It is safe to retry on this function.
func (c *VaultController) DeployVault(v *api.VaultServer) error {
	_, err := c.kubeClient.AppsV1beta1().Deployments(v.Namespace).Get(v.Name, metav1.GetOptions{})
	if !kerrors.IsNotFound(err) {
		glog.Infof("deployment '%s' already exists", v.Name)
		return nil
	}

	saName, err := c.createVaultServiceAccount(v)
	if err != nil {
		return err
	}

	selector := util.LabelsForVault(v.GetName())

	podTempl := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.GetName(),
			Labels:    selector,
			Namespace: v.GetNamespace(),
		},
		Spec: corev1.PodSpec{
			Containers:         []corev1.Container{vaultContainer(v)},
			ServiceAccountName: saName,
			Volumes: []corev1.Volume{{
				Name: VaultConfigVolumeName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: util.ConfigMapNameForVault(v),
						},
					},
				},
			}},
		},
	}

	// Add pod resource policy
	if v.Spec.PodPolicy != nil {
		util.ApplyPodResourcePolicy(&podTempl.Spec, v.Spec.PodPolicy)
	}

	configureVaultServerTLS(&podTempl)

	storageSrv, err := storage.NewStorage(&v.Spec.BackendStorage)
	if err != nil {
		return errors.Wrap(err, "failed to create storage service for vault backend service")
	}
	// add environment variable, volume mount, etc for storage if required
	err = storageSrv.Apply(&podTempl)
	if err != nil {
		return errors.Wrap(err, "failed to apply changes in pod template")
	}

	if v.Spec.Unsealer != nil {
		// add vault unsealer as sidecar
		unseal, err := unsealer.NewUnsealer(v.Spec.Unsealer)
		err = unseal.AddContainer(&podTempl)
		if err != nil {
			return errors.Wrap(err, "failed to add unsealer container")
		}

		// get rbac roles
		rbacRoles := unseal.GetRBAC(v.GetNamespace(), selector)

		//create role and role binding
		err = c.createRoleAndRoleBinding(v, rbacRoles, saName)
		if err != nil {
			return err
		}
	}

	err = c.createVaultDeployment(v, &podTempl)
	if err != nil {
		return err
	}

	err = c.createVaultService(v)
	if err != nil {
		return err
	}

	return nil
}

func (c *VaultController) syncUpgrade(v *api.VaultServer, d *appsv1beta1.Deployment) error {
	// If the deployment version hasn't been updated, roll forward the deployment version
	// but keep the existing active Vault node alive though.
	if d.Spec.Template.Spec.Containers[0].Image != util.VaultImage(v) {
		err := c.UpgradeDeployment(v, d)
		if err != nil {
			return errors.Wrap(err, "unable to upgrade deployment")
		}
	}

	// If there is one active node belonging to the old version, and all other nodes are
	// standby and uptodate, then trigger step-down on active node.
	// It maps to the following conditions on Status:
	// 1. check standby == updated
	// 2. check Available - Updated == Active
	readyToTriggerStepdown := func() bool {
		if len(v.Status.VaultStatus.Active) == 0 {
			return false
		}

		if !reflect.DeepEqual(v.Status.VaultStatus.Standby, v.Status.UpdatedNodes) {
			return false
		}

		ava := append(v.Status.VaultStatus.Standby, v.Status.VaultStatus.Sealed...)
		if !reflect.DeepEqual(ava, v.Status.UpdatedNodes) {
			return false
		}
		return true
	}()

	if readyToTriggerStepdown {
		// This will send SIGTERM to the active Vault pod. It should release HA lock and exit properly.
		// If it failed for some reason, kubelet will send SIGKILL after default grace period (30s) eventually.
		// It take longer but the the lock will get released eventually on failure case.
		err := c.kubeClient.CoreV1().Pods(v.Namespace).Delete(v.Status.VaultStatus.Active, nil)
		if err != nil && !kerrors.IsNotFound(err) {
			return errors.Wrapf(err, "step down: failed to delete active Vault pod (%s): %v", v.Status.VaultStatus.Active)
		}
	}

	return nil
}

// UpgradeDeployment sets deployment spec to:
// - roll forward version
// - keep active Vault node available by setting `maxUnavailable=N-1` and `maxSurge=1`
func (c *VaultController) UpgradeDeployment(v *api.VaultServer, d *appsv1beta1.Deployment) error {
	mu := intstr.FromInt(int(v.Spec.Nodes - 1))

	d, _, err := kutilappsv1beta1.CreateOrPatchDeployment(c.kubeClient, d.ObjectMeta, func(deployment *appsv1beta1.Deployment) *appsv1beta1.Deployment {
		deployment.Spec.Strategy.RollingUpdate.MaxUnavailable = &mu
		deployment.Spec.Template.Spec.Containers[0].Image = util.VaultImage(v)
		return deployment
	})
	if err != nil {
		return errors.Wrapf(err, "failed to upgrade deployment to (%s)", util.VaultImage(v))
	}
	return nil
}

// prepareVaultTLSSecrets creates secret containing following data:
//     ca.crt : <ca.crt-used-to-sign-vault-server-cert>
//     server.crt : <vault-server-cert>
//     server.key : <vault-server-key>
//
// currently used self signed certificate
func (c *VaultController) prepareVaultTLSSecrets(v *api.VaultServer) error {
	_, err := c.kubeClient.CoreV1().Secrets(v.Namespace).Get(VaultTlsSecretName, metav1.GetOptions{})
	if !kerrors.IsNotFound(err) {
		glog.Infof("secret '%s' already exists", VaultTlsSecretName)
		return nil
	} else if !kerrors.IsNotFound(err) {
		return errors.Wrap(err, "vault secret get error")
	}

	store, err := certstore.NewCertStore(afero.NewMemMapFs(), filepath.Join("", "pki"))
	if err != nil {
		return errors.Wrap(err, "certificate store create error")
	}

	err = store.NewCA()
	if err != nil {
		return errors.Wrap(err, "ca certificate create error")
	}

	// ref: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
	altNames := cert.AltNames{
		DNSNames: []string{
			"localhost",
			fmt.Sprintf("*.%s.pod", v.Namespace),
			fmt.Sprintf("%s.%s.svc", v.Name, v.Namespace),
		},
	}

	srvCrt, srvKey, err := store.NewServerCertPair("server", altNames)
	if err != nil {
		return errors.Wrap(err, "vault server create crt/key pair error")
	}

	vaultTlsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   VaultTlsSecretName,
			Labels: util.LabelsForVault(v.Name),
		},
		Data: map[string][]byte{
			CaCertName:     store.CACert(),
			ServerCertName: srvCrt,
			ServerkeyName:  srvKey,
		},
	}

	util.AddOwnerRefToObject(vaultTlsSecret, util.AsOwner(v))

	_, err = c.kubeClient.CoreV1().Secrets(v.Namespace).Create(vaultTlsSecret)
	if err != nil {
		return errors.Wrap(err, "vault tls secret create error")
	}

	return nil
}

// prepareConfig will do:
// - Create listener config
// - Append extra user given config from configMap if user provided it
// - Create backend storage config from backendStorageSpec
// - Create a ConfigMap "${vaultName}-vault-config" containing configuration
func (c *VaultController) prepareConfig(v *api.VaultServer) error {
	_, err := c.kubeClient.CoreV1().ConfigMaps(v.Namespace).Get(util.ConfigMapNameForVault(v), metav1.GetOptions{})
	if !kerrors.IsNotFound(err) {
		glog.Infof("ConfigMap '%s' already exists", util.ConfigMapNameForVault(v))
		return nil
	}
	cfgData := util.GetListenerConfig()

	if len(v.Spec.ConfigMapName) != 0 {
		cm, err := c.kubeClient.CoreV1().ConfigMaps(v.Namespace).Get(v.Spec.ConfigMapName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "get configmap (%s) failed", v.Spec.ConfigMapName)
		}
		cfgData = fmt.Sprintf("%s\n%s", cfgData, cm.Data[filepath.Base(util.VaultConfigFile)])
	}

	storageSrv, err := storage.NewStorage(&v.Spec.BackendStorage)
	if err != nil {
		return errors.Wrap(err, "failed to create storage service for vault backend service")
	}

	storageCfg, err := storageSrv.GetStorageConfig()
	if err != nil {
		return errors.Wrap(err, "create vault storage config failed")
	}
	cfgData = fmt.Sprintf("%s\n%s", cfgData, storageCfg)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   util.ConfigMapNameForVault(v),
			Labels: util.LabelsForVault(v.Name),
		},
		Data: map[string]string{
			filepath.Base(util.VaultConfigFile): cfgData,
		},
	}

	util.AddOwnerRefToObject(cm, util.AsOwner(v))
	_, err = c.kubeClient.CoreV1().ConfigMaps(v.Namespace).Create(cm)
	if err != nil {
		return errors.Wrapf(err, "create new configmap (%s) failed", cm.Name)
	}

	return nil
}

// createVaultDeployment creates deployment for vault
func (c *VaultController) createVaultDeployment(v *api.VaultServer, p *corev1.PodTemplateSpec) error {
	selector := util.LabelsForVault(v.GetName())

	d := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   v.GetName(),
			Labels: selector,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &v.Spec.Nodes,
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: *p,
			Strategy: appsv1beta1.DeploymentStrategy{
				Type: appsv1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1beta1.RollingUpdateDeployment{
					MaxUnavailable: func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
					MaxSurge:       func(a intstr.IntOrString) *intstr.IntOrString { return &a }(intstr.FromInt(1)),
				},
			},
		},
	}

	util.AddOwnerRefToObject(d, util.AsOwner(v))
	_, err := c.kubeClient.AppsV1beta1().Deployments(v.Namespace).Create(d)
	if err != nil {
		return errors.Wrap(err, "unable to create deployments for vault")
	}
	return nil
}

func (c *VaultController) createVaultServiceAccount(v *api.VaultServer) (string, error) {
	selector := util.LabelsForVault(v.GetName())

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.GetName(),
			Namespace: v.GetNamespace(),
			Labels:    selector,
		},
	}

	util.AddOwnerRefToObject(sa, util.AsOwner(v))
	_, err := c.kubeClient.CoreV1().ServiceAccounts(v.GetNamespace()).Create(sa)
	if err != nil {
		return "", errors.Wrap(err, "failed to create service account")
	}

	return sa.GetName(), nil
}

func (c *VaultController) createRoleAndRoleBinding(v *api.VaultServer, roles []rbac.Role, saName string) error {
	selector := util.LabelsForVault(v.GetName())

	for _, role := range roles {
		roleBind := &rbac.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      role.GetName(),
				Namespace: v.GetNamespace(),
				Labels:    selector,
			},
			RoleRef: rbac.RoleRef{
				APIGroup: rbac.GroupName,
				Kind:     "Role",
				Name:     role.GetName(),
			},
			Subjects: []rbac.Subject{
				{
					Kind:      rbac.ServiceAccountKind,
					Name:      saName,
					Namespace: v.GetNamespace(),
				},
			},
		}

		util.AddOwnerRefToObject(role.GetObjectMeta(), util.AsOwner(v))
		_, err := c.kubeClient.RbacV1().Roles(role.GetNamespace()).Create(&role)
		if err != nil {
			return errors.Wrapf(err, "failed to create rbac role(%s)", role.GetName())
		}

		util.AddOwnerRefToObject(roleBind.GetObjectMeta(), util.AsOwner(v))
		_, err = c.kubeClient.RbacV1().RoleBindings(roleBind.GetNamespace()).Create(roleBind)
		if err != nil {
			return errors.Wrapf(err, "failed to create rbac role binding(%s)", roleBind.GetName())
		}
	}

	return nil
}

// createVaultService creates service for vault
func (c *VaultController) createVaultService(v *api.VaultServer) error {
	selector := util.LabelsForVault(v.GetName())

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   v.Name,
			Labels: selector,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
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
		},
	}

	util.AddOwnerRefToObject(svc, util.AsOwner(v))
	_, err := c.kubeClient.CoreV1().Services(v.Namespace).Create(svc)
	if err != nil {
		return errors.Wrap(err, "failed to create vault service")
	}
	return nil
}

func vaultContainer(v *api.VaultServer) corev1.Container {
	return corev1.Container{
		Name:  "vault",
		Image: util.VaultImage(v),
		Command: []string{
			"/bin/vault",
			"server",
			"-config=" + util.VaultConfigFile,
		},
		Env: []corev1.EnvVar{
			{
				Name:  EnvVaultAddr,
				Value: util.VaultServiceURL(v.GetName(), v.GetNamespace(), VaultPort),
			},
			{
				Name:  EnvVaultClusterAddr,
				Value: util.VaultServiceURL(v.GetName(), v.GetNamespace(), VaultClusterPort),
			},
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      VaultConfigVolumeName,
			MountPath: filepath.Dir(util.VaultConfigFile),
		}},
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
	}
}

// TODO : Use user provided certificates
// configureVaultServerTLS mounts the volume containing the vault server TLS assets for the vault pod
func configureVaultServerTLS(pt *corev1.PodTemplateSpec) {
	vaultTLSAssetVolume := "vault-tls-secret"

	pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
		Name: vaultTLSAssetVolume,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: VaultTlsSecretName,
			},
		},
	})

	pt.Spec.Containers[0].VolumeMounts = append(pt.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
		Name:      vaultTLSAssetVolume,
		MountPath: util.VaultTLSAssetDir,
	})
}
