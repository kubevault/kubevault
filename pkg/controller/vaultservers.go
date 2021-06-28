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

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	patchutil "kubevault.dev/apimachinery/client/clientset/versioned/typed/kubevault/v1alpha1/util"
	"kubevault.dev/operator/pkg/eventer"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	apps_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	rbac_util "kmodules.xyz/client-go/rbac/v1"
	"kmodules.xyz/client-go/tools/queue"
)

const (
	VaultPVCName = "data"
)

func (c *VaultController) initVaultServerWatcher() {
	c.vsInformer = c.extInformerFactory.Kubevault().V1alpha1().VaultServers().Informer()
	c.vsQueue = queue.New(api.ResourceKindVaultServer, c.MaxNumRequeues, c.NumThreads, c.runVaultServerInjector)
	c.vsInformer.AddEventHandler(queue.NewReconcilableHandler(c.vsQueue.GetQueue()))
	if c.auditor != nil {
		c.vsInformer.AddEventHandler(c.auditor.ForGVK(api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
	}
	c.vsLister = c.extInformerFactory.Kubevault().V1alpha1().VaultServers().Lister()
}

// runVaultSeverInjector gets the vault server object indexed by the key from cache
// and initializes, reconciles or garbage collects the vault cluster as needed.
func (c *VaultController) runVaultServerInjector(key string) error {
	obj, exists, err := c.vsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a VaultServer, so that we will see a delete for one d
		klog.Warningf("VaultServer %s does not exist anymore\n", key)

		// stop vault status monitor
		if ctxWithCancel, ok := c.ctxCancels[key]; ok {
			ctxWithCancel.Cancel()
			delete(c.ctxCancels, key)
		}

		// stop auth method controller go routine if have any
		if ctxWithCancel, ok := c.authMethodCtx[key]; ok {
			ctxWithCancel.Cancel()
			delete(c.authMethodCtx, key)
		}

	} else {
		vs := obj.(*api.VaultServer).DeepCopy()

		klog.Infof("Sync/Add/Update for VaultServer %s/%s\n", vs.Namespace, vs.Name)

		if vs.DeletionTimestamp != nil {
			return nil
		} else {
			v, err := NewVault(vs, c.clientConfig, c.kubeClient, c.extClient)
			if err != nil {
				return errors.Wrapf(err, "for VaultServer %s/%s", vs.Namespace, vs.Name)
			}

			err = c.reconcileVault(vs, v)
			if err != nil {
				return errors.Wrapf(err, "for VaultServer %s/%s", vs.Namespace, vs.Name)
			}
		}
	}
	return nil
}

// reconcileVault reconciles the vault cluster's state to the spec specified by v
// by preparing the TLS secrets, deploying vault cluster,
// and finally updating the vault deployment if needed.
// It also creates AppBinding containing vault connection configuration
func (c *VaultController) reconcileVault(vs *api.VaultServer, v Vault) error {
	err := c.CreateVaultTLSSecret(vs, v)
	if err != nil {
		_, err2 := patchutil.UpdateVaultServerStatus(
			context.TODO(),
			c.extClient.KubevaultV1alpha1(),
			vs.ObjectMeta,
			func(status *api.VaultServerStatus) *api.VaultServerStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToCreateVaultTLSSecret",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create vault server tls secret")})
	}

	err = c.CreateVaultConfig(vs, v)
	if err != nil {
		_, err2 := patchutil.UpdateVaultServerStatus(
			context.TODO(),
			c.extClient.KubevaultV1alpha1(),
			vs.ObjectMeta,
			func(status *api.VaultServerStatus) *api.VaultServerStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToCreateVaultConfig",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to create vault config")})
	}

	err = c.DeployVault(vs, v)
	if err != nil {
		_, err2 := patchutil.UpdateVaultServerStatus(
			context.TODO(),
			c.extClient.KubevaultV1alpha1(),
			vs.ObjectMeta,
			func(status *api.VaultServerStatus) *api.VaultServerStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToDeployVault",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to deploy vault")})
	}

	err = c.ensureAppBindings(vs, v)
	if err != nil {
		_, err2 := patchutil.UpdateVaultServerStatus(
			context.TODO(),
			c.extClient.KubevaultV1alpha1(),
			vs.ObjectMeta,
			func(status *api.VaultServerStatus) *api.VaultServerStatus {
				status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
					Type:    kmapi.ConditionFailed,
					Status:  core.ConditionTrue,
					Reason:  "FailedToCreateAppBinding",
					Message: err.Error(),
				})
				return status
			},
			metav1.UpdateOptions{},
		)
		return utilerrors.NewAggregate([]error{err2, errors.Wrap(err, "failed to deploy vault")})
	}

	_, err = patchutil.UpdateVaultServerStatus(
		context.TODO(),
		c.extClient.KubevaultV1alpha1(),
		vs.ObjectMeta,
		func(status *api.VaultServerStatus) *api.VaultServerStatus {
			status.ObservedGeneration = vs.Generation
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.Condition{
				Type:    kmapi.ConditionReady,
				Status:  core.ConditionTrue,
				Reason:  "Provisioned",
				Message: "vault server is ready to use",
			})
			return status
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		return errors.Wrap(err, "failed to update status")
	}

	// Add vault monitor to watch vault seal or unseal status
	key := vs.GetKey()
	if _, ok := c.ctxCancels[key]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		c.ctxCancels[key] = CtxWithCancel{
			Ctx:    ctx,
			Cancel: cancel,
		}
		go c.monitorAndUpdateStatus(ctx, vs)
	}

	// Run auth method reconcile
	c.runAuthMethodsReconcile(vs)

	return nil
}

func (c *VaultController) CreateVaultTLSSecret(vs *api.VaultServer, v Vault) error {
	if err := v.EnsureCA(); err != nil {
		return err
	}
	if err := v.EnsureServerTLS(); err != nil {
		return err
	}
	if err := v.EnsureClientTLS(); err != nil {
		return err
	}
	if err := v.EnsureStorageTLS(); err != nil {
		return err
	}
	return nil
}

func (c *VaultController) CreateVaultConfig(vs *api.VaultServer, v Vault) error {
	cm, err := v.GetConfig()
	if err != nil {
		return err
	}
	return ensureConfigSecret(c.kubeClient, vs, cm)
}

// - create service account for vault pod
// - create deployment
// - create service
// - create rbac role, rolebinding and cluster rolebinding
func (c *VaultController) DeployVault(vs *api.VaultServer, v Vault) error {
	saList := v.GetServiceAccounts()
	for _, sa := range saList {
		err := ensureServiceAccount(c.kubeClient, vs, &sa)
		if err != nil {
			return err
		}
	}

	svc := v.GetService()
	err := ensureService(c.kubeClient, vs, svc)
	if err != nil {
		return err
	}

	rList, rBList := v.GetRBACRolesAndRoleBindings()
	err = ensureRoleAndRoleBinding(c.kubeClient, vs, rList, rBList)
	if err != nil {
		return err
	}

	cRB := v.GetRBACClusterRoleBinding()
	err = ensureClusterRoleBinding(c.kubeClient, vs, cRB)
	if err != nil {
		return err
	}

	// apply changes to PodTemplate after creating service accounts
	// because unsealer use token reviewer jwt to enable kubernetes auth

	podT := v.GetPodTemplate(v.GetContainer(), vs.ServiceAccountName())
	err = v.Apply(podT)
	if err != nil {
		return err
	}

	serviceName := vs.ServiceName(api.VaultServerServiceInternal)
	headlessSvc := v.GetGoverningService()
	err = ensureService(c.kubeClient, vs, headlessSvc)
	if err != nil {
		return err
	}

	// XXX Add pvc support
	claims := c.getPVCs(vs)
	sts := v.GetStatefulSet(serviceName, podT, claims)
	err = ensureStatefulSet(c.kubeClient, vs, sts)
	if err != nil {
		return err
	}

	if vs.Spec.Monitor != nil && vs.Spec.Monitor.Prometheus != nil {
		if _, vt, err := c.ensureStatsService(vs); err != nil { // Error ignored intentionally
			c.recorder.Eventf(
				vs,
				core.EventTypeWarning,
				eventer.EventReasonStatsServiceReconcileFailed,
				"Failed to ensure stats Service %s. Reason: %v",
				vs.StatsServiceName(),
				err,
			)
		} else if vt != kutil.VerbUnchanged {
			c.recorder.Eventf(
				vs,
				core.EventTypeNormal,
				eventer.EventReasonStatsServiceReconcileSuccessful,
				"Successfully %s stats Service %s",
				vt,
				vs.StatsServiceName(),
			)
		}
	} else {
		if err := c.ensureStatsServiceDeleted(vs); err != nil { // Error ignored intentionally
			klog.Warningf("failed to delete stats Service %s, reason: %s", vs.StatsServiceName(), err)
		} else {
			c.recorder.Eventf(
				vs,
				core.EventTypeNormal,
				eventer.EventReasonStatsServiceDeleteSuccessful,
				"Successfully deleted stats Service %s",
				vs.StatsServiceName(),
			)
		}
	}

	if err = c.manageMonitor(vs); err != nil {
		return err
	}
	return nil
}

func (c *VaultController) getPVCs(vs *api.VaultServer) []core.PersistentVolumeClaim {
	if vs.Spec.Backend.Raft != nil && vs.Spec.Backend.Raft.Storage != nil {
		pvc := core.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: VaultPVCName,
			},
			Spec: *vs.Spec.Backend.Raft.Storage,
		}

		if len(pvc.Spec.AccessModes) == 0 {
			pvc.Spec.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
		}

		if pvc.Spec.StorageClassName != nil {
			pvc.Annotations = map[string]string{
				"volume.beta.kubernetes.io/storage-class": *pvc.Spec.StorageClassName,
			}
		}
		return []core.PersistentVolumeClaim{pvc}
	}
	return nil
}

// ensureServiceAccount creates/patches service account
func ensureServiceAccount(kc kubernetes.Interface, vs *api.VaultServer, sa *core.ServiceAccount) error {
	_, _, err := core_util.CreateOrPatchServiceAccount(context.TODO(), kc, sa.ObjectMeta, func(in *core.ServiceAccount) *core.ServiceAccount {
		in.Labels = core_util.UpsertMap(in.Labels, sa.Labels)
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in
	}, metav1.PatchOptions{})
	return err
}

// ensureStatefulSet creates/patches sts
func ensureStatefulSet(kc kubernetes.Interface, vs *api.VaultServer, sts *appsv1.StatefulSet) error {
	_, _, err := apps_util.CreateOrPatchStatefulSet(context.TODO(), kc, sts.ObjectMeta, func(in *appsv1.StatefulSet) *appsv1.StatefulSet {
		in.Labels = core_util.UpsertMap(in.Labels, sts.Labels)
		in.Annotations = core_util.UpsertMap(in.Annotations, sts.Annotations)
		in.Spec.Replicas = sts.Spec.Replicas
		in.Spec.Selector = sts.Spec.Selector
		in.Spec.ServiceName = sts.Spec.ServiceName
		in.Spec.UpdateStrategy = sts.Spec.UpdateStrategy

		in.Spec.Template.Labels = sts.Spec.Template.Labels
		in.Spec.Template.Annotations = sts.Spec.Template.Annotations
		in.Spec.Template.Spec.Containers = core_util.UpsertContainers(in.Spec.Template.Spec.Containers, sts.Spec.Template.Spec.Containers)
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(in.Spec.Template.Spec.InitContainers, sts.Spec.Template.Spec.InitContainers)
		in.Spec.Template.Spec.ServiceAccountName = sts.Spec.Template.Spec.ServiceAccountName
		in.Spec.Template.Spec.NodeSelector = sts.Spec.Template.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = sts.Spec.Template.Spec.Affinity
		if sts.Spec.Template.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = sts.Spec.Template.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = sts.Spec.Template.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = sts.Spec.Template.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = sts.Spec.Template.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = sts.Spec.Template.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = sts.Spec.Template.Spec.SecurityContext
		in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, sts.Spec.Template.Spec.Volumes...)
		in.Spec.VolumeClaimTemplates = sts.Spec.VolumeClaimTemplates

		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in

	}, metav1.PatchOptions{})
	return err
}

// ensureService creates/patches service
func ensureService(kc kubernetes.Interface, vs *api.VaultServer, svc *core.Service) error {
	_, _, err := core_util.CreateOrPatchService(context.TODO(), kc, svc.ObjectMeta, func(in *core.Service) *core.Service {
		in.Labels = core_util.UpsertMap(in.Labels, svc.Labels)
		in.Annotations = core_util.UpsertMap(in.Annotations, svc.Annotations)

		in.Spec.Selector = svc.Spec.Selector
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, svc.Spec.Ports)
		if svc.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = svc.Spec.ClusterIP
		}
		if svc.Spec.Type != "" {
			in.Spec.Type = svc.Spec.Type
		}
		if svc.Spec.LoadBalancerIP != "" {
			in.Spec.LoadBalancerIP = svc.Spec.LoadBalancerIP
		}
		in.Spec.ExternalIPs = svc.Spec.ExternalIPs
		in.Spec.LoadBalancerSourceRanges = svc.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = svc.Spec.ExternalTrafficPolicy
		if svc.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = svc.Spec.HealthCheckNodePort
		}
		if svc.Spec.PublishNotReadyAddresses {
			in.Spec.PublishNotReadyAddresses = true
		}
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in
	}, metav1.PatchOptions{})
	return err
}

// ensureRoleAndRoleBinding creates or patches rbac role and rolebinding
func ensureRoleAndRoleBinding(kc kubernetes.Interface, vs *api.VaultServer, roles []rbac.Role, rBindings []rbac.RoleBinding) error {
	for _, role := range roles {
		_, _, err := rbac_util.CreateOrPatchRole(context.TODO(), kc, role.ObjectMeta, func(in *rbac.Role) *rbac.Role {
			in.Labels = core_util.UpsertMap(in.Labels, role.Labels)
			in.Annotations = core_util.UpsertMap(in.Annotations, role.Annotations)
			in.Rules = role.Rules
			core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create rbac role %s/%s", role.Namespace, role.Name)
		}
	}

	for _, rb := range rBindings {
		_, _, err := rbac_util.CreateOrPatchRoleBinding(context.TODO(), kc, rb.ObjectMeta, func(in *rbac.RoleBinding) *rbac.RoleBinding {
			in.Labels = core_util.UpsertMap(in.Labels, rb.Labels)
			in.RoleRef = rb.RoleRef
			in.Subjects = rb.Subjects
			core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
			return in
		}, metav1.PatchOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create rbac role binding %s/%s", rb.Namespace, rb.Name)
		}
	}
	return nil
}

// ensureConfigSecret creates/patches Secret
func ensureConfigSecret(kc kubernetes.Interface, vs *api.VaultServer, secret *core.Secret) error {
	_, _, err := core_util.CreateOrPatchSecret(context.TODO(), kc, secret.ObjectMeta, func(in *core.Secret) *core.Secret {
		in.Labels = core_util.UpsertMap(in.Labels, secret.Labels)
		in.Annotations = core_util.UpsertMap(in.Annotations, secret.Annotations)
		in.Data = secret.Data
		in.StringData = secret.StringData
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in
	}, metav1.PatchOptions{})
	return err
}

// ensureClusterRoleBinding creates or patches rbac ClusterRoleBinding
func ensureClusterRoleBinding(kc kubernetes.Interface, vs *api.VaultServer, cRBinding rbac.ClusterRoleBinding) error {
	_, _, err := rbac_util.CreateOrPatchClusterRoleBinding(
		context.TODO(),
		kc,
		cRBinding.ObjectMeta,
		func(in *rbac.ClusterRoleBinding) *rbac.ClusterRoleBinding {
			in.Labels = core_util.UpsertMap(in.Labels, cRBinding.Labels)
			in.RoleRef = cRBinding.RoleRef
			in.Subjects = cRBinding.Subjects
			core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
			return in
		},
		metav1.PatchOptions{},
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create rbac role %s/%s", cRBinding.Namespace, cRBinding.Name)
	}
	return nil
}
