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
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	vaultconfig "kubevault.dev/apimachinery/apis/config/v1alpha1"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	policyapi "kubevault.dev/apimachinery/apis/policy/v1alpha1"
	vaultcs "kubevault.dev/apimachinery/client/clientset/versioned/typed/kubevault/v1alpha1"
	patchutil "kubevault.dev/apimachinery/client/clientset/versioned/typed/kubevault/v1alpha1/util"
	policycs "kubevault.dev/apimachinery/client/clientset/versioned/typed/policy/v1alpha1"
	policyutil "kubevault.dev/apimachinery/client/clientset/versioned/typed/policy/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault"
	authtype "kubevault.dev/operator/pkg/vault/auth/types"

	"github.com/golang/glog"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	core_util "kmodules.xyz/client-go/core/v1"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

const policyForAuthController = `
path "sys/auth" {
  capabilities = ["read", "list", ]
}

path "sys/auth/*" {
  capabilities = ["sudo", "create", "read", "update", "delete"]
}
`

const (
	ttlForAuthMethod = "24h"
)

func (c *VaultController) runAuthMethodsReconcile(vs *api.VaultServer) {
	if vs == nil {
		glog.Errorln("VaultServer is nil")
		return
	}

	key := vs.GetKey()
	ctx, cancel := context.WithCancel(context.Background())
	ctxCancel := CtxWithCancel{
		Ctx:    ctx,
		Cancel: cancel,
	}

	if ctx, ok := c.authMethodCtx[key]; ok {
		// stop previous infinitely running go routine if have any
		ctx.Cancel()
	}

	// run a new go routine for updated auth methods
	c.authMethodCtx[key] = ctxCancel
	go c.reconcileAuthMethods(vs, ctxCancel.Ctx)
}

// tasks:
//	- create VaultPolicy and VaultPolicyBinding, it will not create those until vault is ready
//  - enable or disable auth methods in vault
func (c *VaultController) reconcileAuthMethods(vs *api.VaultServer, ctx context.Context) {
	if vs == nil {
		glog.Errorf("VaultServer is nil")
		return
	}

	var err error
	vs, err = waitUntilVaultServerIsReady(c.extClient.KubevaultV1alpha1(), vs, ctx.Done())
	if err != nil {
		glog.Errorf("error when wating for VaultServer to get ready: %s", err)
		return
	}

	vp := vaultPolicyForAuthMethod(vs)
	err = ensureVaultPolicy(c.extClient.PolicyV1alpha1(), vp, vs)
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}
	// wait until VaultPolicy is succeeded
	err = waitUntilVaultPolicyIsReady(c.extClient.PolicyV1alpha1(), vp, ctx.Done())
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}

	// ensure VaultPolicyBinding
	vpb := vaultPolicyBindingForAuthMethod(vs)
	err = ensureVaultPolicyBinding(c.extClient.PolicyV1alpha1(), vpb, vs)
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}
	// wait until VaultPolicyBinding is succeeded
	err = waitUntilVaultPolicyBindingIsReady(c.extClient.PolicyV1alpha1(), vpb, ctx.Done())
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}

	// enable or disable auth method based on .spec.authMethods and .status.authMethodStatus
	vc, err := newVaultClientForAuthMethodController(c.kubeClient, c.appCatalogClient, vs)
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}

	authStatus, err := enableAuthMethods(vc, vs.Spec.AuthMethods)
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}

	authDisableStatus := disableAuthMethods(vc, vs.Spec.AuthMethods, vs.Status.AuthMethodStatus)
	authStatus = append(authStatus, authDisableStatus...)

	_, err = patchutil.UpdateVaultServerStatus(
		context.TODO(),
		c.extClient.KubevaultV1alpha1(),
		vs.ObjectMeta,
		func(status *api.VaultServerStatus) *api.VaultServerStatus {
			status.AuthMethodStatus = authStatus
			return status
		},
		metav1.UpdateOptions{},
	)
	if err != nil {
		glog.Errorf("auth method controller: for VaultServer %s/%s: %s", vs.Namespace, vs.Name, err)
		return
	}

	glog.Infof("auth method controller: for VaultServer %s/%s: auth method enable or disable operation applied", vs.Namespace, vs.Name)
}

func vaultPolicyForAuthMethod(vs *api.VaultServer) *policyapi.VaultPolicy {
	policy := &policyapi.VaultPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vs.PolicyNameForAuthMethodController(),
			Namespace: vs.Namespace,
			Labels:    vs.OffshootLabels(),
		},
		Spec: policyapi.VaultPolicySpec{
			VaultRef: core.LocalObjectReference{
				Name: vs.AppBindingName(),
			},
			PolicyDocument: policyForAuthController,
		},
	}
	return policy
}

func vaultPolicyBindingForAuthMethod(vs *api.VaultServer) *policyapi.VaultPolicyBinding {
	var pb = &policyapi.VaultPolicyBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vs.PolicyNameForAuthMethodController(),
			Namespace: vs.Namespace,
			Labels:    vs.OffshootLabels(),
		},
		Spec: policyapi.VaultPolicyBindingSpec{
			VaultRef: core.LocalObjectReference{Name: vs.AppBindingName()},
			Policies: []policyapi.PolicyIdentifier{
				{
					Ref: vs.PolicyNameForAuthMethodController(),
				},
			},
			SubjectRef: policyapi.SubjectRef{
				Kubernetes: &policyapi.KubernetesSubjectRef{
					Path:                     string(api.AuthTypeKubernetes),
					ServiceAccountNames:      []string{vs.ServiceAccountName()},
					ServiceAccountNamespaces: []string{vs.Namespace},
					TTL:                      ttlForAuthMethod,
					MaxTTL:                   ttlForAuthMethod,
					Period:                   ttlForAuthMethod,
				},
			},
		},
	}
	return pb
}

func ensureVaultPolicy(c policycs.PolicyV1alpha1Interface, vp *policyapi.VaultPolicy, vs *api.VaultServer) error {
	_, _, err := policyutil.CreateOrPatchVaultPolicy(context.TODO(), c, vp.ObjectMeta, func(in *policyapi.VaultPolicy) *policyapi.VaultPolicy {
		in.Labels = core_util.UpsertMap(in.Labels, vp.Labels)
		in.Spec.PolicyDocument = vp.Spec.PolicyDocument
		in.Spec.Policy = vp.Spec.Policy
		in.Spec.VaultRef = vp.Spec.VaultRef
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to ensure VaultPolicy %s/%s", vp.Namespace, vp.Name)
	}
	return nil
}

func ensureVaultPolicyBinding(c policycs.PolicyV1alpha1Interface, vpb *policyapi.VaultPolicyBinding, vs *api.VaultServer) error {
	_, _, err := policyutil.CreateOrPatchVaultPolicyBinding(context.TODO(), c, vpb.ObjectMeta, func(in *policyapi.VaultPolicyBinding) *policyapi.VaultPolicyBinding {
		in.Labels = core_util.UpsertMap(in.Labels, vpb.Labels)
		in.Spec = vpb.Spec
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))
		return in
	}, metav1.PatchOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to ensure VaultPolicyBinding %s/%s", vpb.Namespace, vpb.Name)
	}
	return nil
}

func enableAuthMethods(vc *vaultapi.Client, auths []api.AuthMethod) ([]api.AuthMethodStatus, error) {
	// in auth list path will always be appended with '/'
	authList, err := vc.Sys().ListAuth()
	if err != nil {
		return nil, err
	}

	var resp []api.AuthMethodStatus

	for _, au := range auths {
		p := filepath.Clean(au.Path) + "/"

		if got, ok := authList[p]; ok {
			// auth method already enabled in this path
			if got.Type != au.Type {
				resp = append(resp, api.AuthMethodStatus{
					Type:   au.Type,
					Path:   au.Path,
					Status: api.AuthMethodEnableFailed,
					Reason: fmt.Sprintf("%s type auth already enabled in this path", got.Type),
				})
			} else {
				resp = append(resp, api.AuthMethodStatus{
					Type:   au.Type,
					Path:   au.Path,
					Status: api.AuthMethodEnableSucceeded,
					Reason: "",
				})
			}
		} else {
			// auth method is not enabled in this path
			opts := &vaultapi.EnableAuthOptions{
				Type:        au.Type,
				Description: au.Description,
				PluginName:  au.PluginName,
				Local:       au.Local,
			}
			if au.Config != nil {
				cf := au.Config
				opts.Config = vaultapi.AuthConfigInput{
					DefaultLeaseTTL:           cf.DefaultLeaseTTL,
					MaxLeaseTTL:               cf.MaxLeaseTTL,
					PluginName:                cf.PluginName,
					AuditNonHMACRequestKeys:   cf.AuditNonHMACRequestKeys,
					AuditNonHMACResponseKeys:  cf.AuditNonHMACResponseKeys,
					ListingVisibility:         cf.ListingVisibility,
					PassthroughRequestHeaders: cf.PassthroughRequestHeaders,
				}
			}

			err = vc.Sys().EnableAuthWithOptions(au.Path, opts)
			if err != nil {
				resp = append(resp, api.AuthMethodStatus{
					Type:   au.Type,
					Path:   au.Path,
					Status: api.AuthMethodEnableFailed,
					Reason: err.Error(),
				})
			} else {
				resp = append(resp, api.AuthMethodStatus{
					Type:   au.Type,
					Path:   au.Path,
					Status: api.AuthMethodEnableSucceeded,
					Reason: "",
				})
			}
		}
	}
	return resp, nil
}

// Disable auth methods that are not in the 'expected' auth methods but in the 'has' auth methods
// returns the auth methods that are failed to disable
func disableAuthMethods(vc *vaultapi.Client, expected []api.AuthMethod, has []api.AuthMethodStatus) []api.AuthMethodStatus {
	authMap := map[string]bool{}
	for _, au := range expected {
		authMap[filepath.Clean(au.Path)] = true
	}

	var failedToDisable []api.AuthMethodStatus
	for _, au := range has {
		p := filepath.Clean(au.Path)
		if ok := authMap[p]; !ok && au.Status == api.AuthMethodEnableSucceeded {
			err := vc.Sys().DisableAuth(p)
			if err != nil {
				failedToDisable = append(failedToDisable, api.AuthMethodStatus{
					Path:   au.Path,
					Type:   au.Type,
					Status: api.AuthMethodDisableFailed,
					Reason: err.Error(),
				})
			}
		}
	}
	return failedToDisable
}

// waitUntilVaultServerIsReady will wait until vault server is ready.
// If it's ready, then it will return updated VaultServer.
// If it's not found, then it will return error
func waitUntilVaultServerIsReady(c vaultcs.KubevaultV1alpha1Interface, vs *api.VaultServer, stopCh <-chan struct{}) (*api.VaultServer, error) {
	var err error
	attempt := 0
	err = wait.PollUntil(5*time.Second, func() (done bool, err error) {
		attempt++
		var err2 error
		vs, err2 = c.VaultServers(vs.Namespace).Get(context.TODO(), vs.Name, metav1.GetOptions{})
		if err2 != nil {
			return false, err2
		}

		if vs.Status.Phase == api.ClusterPhaseRunning {
			return true, nil
		}

		glog.Infof("auth method controller: attempt %d: VaultServer %s/%s is not ready", attempt, vs.Namespace, vs.Name)
		return false, nil
	}, stopCh)
	return vs, err
}

// waitUntilVaultPolicyIsReady will wait until VaultPolicy is ready.
func waitUntilVaultPolicyIsReady(c policycs.PolicyV1alpha1Interface, vp *policyapi.VaultPolicy, stopCh <-chan struct{}) error {
	var err error
	attempt := 0
	err = wait.PollUntil(2*time.Second, func() (done bool, err error) {
		attempt++
		var err2 error
		vp, err2 = c.VaultPolicies(vp.Namespace).Get(context.TODO(), vp.Name, metav1.GetOptions{})
		if err2 != nil {
			return false, err2
		}

		if vp.Status.Phase == policyapi.PolicySuccess {
			return true, nil
		}

		glog.Infof("auth method controller: attempt %d: VaultPolicy %s/%s is not succeeded", attempt, vp.Namespace, vp.Name)
		return false, nil
	}, stopCh)
	return err
}

// waitUntilVaultPolicyBindingIsReady will wait until VaultPolicyBinding is ready.
func waitUntilVaultPolicyBindingIsReady(c policycs.PolicyV1alpha1Interface, vpb *policyapi.VaultPolicyBinding, stopCh <-chan struct{}) error {
	var err error
	attempt := 0
	err = wait.PollUntil(2*time.Second, func() (done bool, err error) {
		attempt++
		var err2 error
		vpb, err2 = c.VaultPolicyBindings(vpb.Namespace).Get(context.TODO(), vpb.Name, metav1.GetOptions{})
		if err2 != nil {
			return false, err2
		}

		if vpb.Status.Phase == policyapi.PolicyBindingSuccess {
			return true, nil
		}

		glog.Infof("auth method controller: attempt %d: VaultPolicyBinding %s/%s is not succeeded", attempt, vpb.Namespace, vpb.Name)
		return false, nil
	}, stopCh)
	return err
}

func newVaultClientForAuthMethodController(kc kubernetes.Interface, appc appcat_cs.AppcatalogV1alpha1Interface, vs *api.VaultServer) (*vaultapi.Client, error) {
	conf, err := json.Marshal(vaultconfig.VaultServerConfiguration{
		Kubernetes: &vaultconfig.KubernetesAuthConfig{
			ServiceAccountName: vs.ServiceAccountName(),
		},
		VaultRole: vaultPolicyBindingForAuthMethod(vs).PolicyBindingName(),
	})
	if err != nil {
		return nil, err
	}

	vApp, err := appc.AppBindings(vs.Namespace).Get(context.TODO(), vs.AppBindingName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	vApp.Spec.Parameters = &runtime.RawExtension{
		Raw: conf,
	}

	authInfo, err := authtype.GetAuthInfoFromAppBinding(kc, vApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authentication information")
	}

	return vault.NewClientWithAppBinding(kc, authInfo)
}
