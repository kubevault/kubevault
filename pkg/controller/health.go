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
	"sync"
	"time"

	"kubevault.dev/apimachinery/apis"
	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	cs_util "kubevault.dev/apimachinery/client/clientset/versioned/typed/kubevault/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/util"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	HealthCheckInterval = 10 * time.Second
	ContextTimeOut      = 30 * time.Second
)

func (c *VaultController) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckVaultServerHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckVaultServerHealth(stopCh)
}

func (c *VaultController) CheckVaultServerHealth(stopCh <-chan struct{}) {
	klog.Info("Starting Vault Server health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down Vault Server health checker...")
			return
		default:
			c.CheckVaultServerHealthOnce()
			time.Sleep(HealthCheckInterval)
		}
	}
}

func (c *VaultController) CheckVaultServerHealthOnce() {
	vsList, err := c.vsLister.VaultServers(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list Vault Server objects with: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	for idx := range vsList {
		vs := vsList[idx]

		// If the VS object is deleted or halted, no need to perform health check.
		if vs.DeletionTimestamp != nil || vs.Spec.Halted {
			continue
		}

		wg.Add(1)
		go func(vs *api.VaultServer) {
			defer func() {
				wg.Done()
			}()

			ctx, cancel := context.WithTimeout(context.Background(), ContextTimeOut)
			defer cancel()

			vaultClient, err := c.getVaultServiceSpecificClient(vs)
			if err != nil {
				klog.Errorf("failed generating Client for Vault Service with %s", err.Error())
				return
			}

			// Todo:  make the health check call using the vaultClient
			hr, err := vaultClient.Sys().Health()
			conditions := vs.Status.Conditions

			// Todo: Update conditions when Health Response is Nil or Error
			if err != nil || hr == nil {
				_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
					func(in *api.VaultServerStatus) *api.VaultServerStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    apis.VaultServerAcceptingConnection,
								Status:  core.ConditionFalse,
								Message: "VaultServer not initialized yet",
								Reason:  "VaultServerNotInitialized",
							})
						return in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
					return
				}

				return
			}

			// Todo: Update conditions as Health Response is Not Nil

			// Todo: Update conditions If Initialized
			if hr.Initialized {
				// Todo:
				//  - Initializing must be False
				//  - Initialized must be True
				//  - AcceptingConnection must be True

				// Todo: Make Initializing False
				if !kmapi.IsConditionFalse(conditions, apis.VaultServerInitializing) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerInitializing,
									Status:  core.ConditionFalse,
									Message: "VaultServer Initializing process is completed",
									Reason:  "VaultServerInitializingCompleted",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}

				// Todo: Make Initialized True
				if !kmapi.IsConditionTrue(conditions, apis.VaultServerInitialized) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerInitialized,
									Status:  core.ConditionTrue,
									Message: "VaultServer is already Initialized",
									Reason:  "VaultServerInitialized",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}

				// Todo: Make AcceptingConnection True
				if !kmapi.IsConditionTrue(conditions, apis.VaultServerAcceptingConnection) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerAcceptingConnection,
									Status:  core.ConditionTrue,
									Message: "VaultServer is initialized and accepting connection",
									Reason:  "VaultServerAcceptingConnection",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}
			}

			// Todo: Update conditions for Initialized but Sealed
			if hr.Initialized && hr.Sealed {
				// Todo:
				//  - Unsealing Must be True
				//  - Unsealed Must be False

				// Todo: Make Unsealing True
				if !kmapi.IsConditionTrue(conditions, apis.VaultServerUnsealing) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerUnsealing,
									Status:  core.ConditionTrue,
									Message: "VaultServer is initialized and started unsealing",
									Reason:  "VaultServerUnsealing",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}

				// Todo: Make Unsealed False
				if !kmapi.IsConditionFalse(conditions, apis.VaultServerUnsealed) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerUnsealed,
									Status:  core.ConditionFalse,
									Message: "VaultServer unsealing is not completed",
									Reason:  "VaultServerUnsealed",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}
			}

			// Todo: Update conditions for Initialized and Unsealed
			if hr.Initialized && !hr.Sealed {
				// Todo:
				//  - Unsealing Must be False
				//  - Unsealed Must be True

				// Todo: Make Unsealing False
				if !kmapi.IsConditionFalse(conditions, apis.VaultServerUnsealing) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerUnsealing,
									Status:  core.ConditionFalse,
									Message: "VaultServer unsealing is completed",
									Reason:  "VaultServerUnsealingCompleted",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}

				// Todo: Make Unsealed True
				if !kmapi.IsConditionTrue(conditions, apis.VaultServerUnsealed) {
					_, err = cs_util.UpdateVaultServerStatus(ctx, c.extClient.KubevaultV1alpha1(), vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    apis.VaultServerUnsealed,
									Status:  core.ConditionTrue,
									Message: "VaultServer is initialized and unsealed",
									Reason:  "VaultServerUnsealed",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}
			}
		}(vs)
	}
	// Wait until all go-routine complete executions
	wg.Wait()
}

func (c *VaultController) getVaultServiceSpecificClient(vs *api.VaultServer) (*vaultapi.Client, error) {
	tlsConfig := &vaultapi.TLSConfig{
		Insecure: true,
	}

	svcURL := fmt.Sprintf("%s://%s.%s.svc:%d", vs.Scheme(), vs.ServiceName(api.VaultServerServiceVault), vs.Namespace, apis.VaultAPIPort)

	vaultClient, err := util.NewVaultClient(svcURL, tlsConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating client for the vault service (%s/%s).", vs.Namespace, vs.ServiceName(api.VaultServerServiceVault))
	}

	return vaultClient, nil
}
