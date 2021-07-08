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

	conapi "kubevault.dev/apimachinery/apis"
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

			vaultClient, err := c.getVaultServiceSpecificClient(vs)
			if err != nil {
				klog.Errorf("====== failed generating Client for Vault Service with %s ======", err.Error())
				return
			}

			// Todo:  make the health check call using the vaultClient
			hr, err := vaultClient.Sys().Health()

			if hr != nil {
				klog.Infof("=========== initialized: %v, sealed: %v ============", hr.Initialized, hr.Sealed)
			}

			if err != nil || hr == nil {
				klog.Warningf(" =================== error or hr nil ================= %s", err)
				_, err = cs_util.UpdateVaultServerStatus(
					context.TODO(),
					c.extClient.KubevaultV1alpha1(),
					vs.ObjectMeta,
					func(in *api.VaultServerStatus) *api.VaultServerStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    conapi.VaultServerInitializing,
								Status:  core.ConditionTrue,
								Message: "",
								Reason:  "",
							})
						return in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					klog.Errorf("Failed to update Initializing to True for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
					return
				}

				_, err = cs_util.UpdateVaultServerStatus(
					context.TODO(),
					c.extClient.KubevaultV1alpha1(),
					vs.ObjectMeta,
					func(in *api.VaultServerStatus) *api.VaultServerStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    conapi.VaultServerUnsealed,
								Status:  core.ConditionFalse,
								Message: "",
								Reason:  "",
							})
						return in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					klog.Errorf("Failed to update Unsealed to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
					return
				}
			} else {
				klog.Info("========================= success in requesting health info =======================")
				//i. 200 if initialized, unsealed, and active
				//ii. 429 if unsealed and standby
				//iii. 472 if disaster recovery mode replication secondary and active
				//iv. 473 if performance standby
				//v. 501 if not initialized
				//vi. 503 if sealed
				// Todo: Initializing is False
				_, err = cs_util.UpdateVaultServerStatus(
					context.TODO(),
					c.extClient.KubevaultV1alpha1(),
					vs.ObjectMeta,
					func(in *api.VaultServerStatus) *api.VaultServerStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:    conapi.VaultServerInitializing,
								Status:  core.ConditionFalse,
								Message: "",
								Reason:  "",
							})
						return in
					},
					metav1.UpdateOptions{},
				)
				if err != nil {
					klog.Errorf("Failed to update Initializing to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
					return
				}

				// Todo: Check Health Response and Update conditions array properly

				// Todo: ======================= Check for Initialized condition ==============================
				if hr.Initialized {
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerInitialized,
									Status:  core.ConditionTrue,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Initialized to True for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}

					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerUnsealing,
									Status:  core.ConditionTrue,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Unsealing to True for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}

					if hr.Sealed {
						_, err = cs_util.UpdateVaultServerStatus(
							context.TODO(),
							c.extClient.KubevaultV1alpha1(),
							vs.ObjectMeta,
							func(in *api.VaultServerStatus) *api.VaultServerStatus {
								in.Conditions = kmapi.SetCondition(in.Conditions,
									kmapi.Condition{
										Type:    conapi.VaultServerUnsealing,
										Status:  core.ConditionTrue,
										Message: "",
										Reason:  "",
									})
								return in
							},
							metav1.UpdateOptions{},
						)
						if err != nil {
							klog.Errorf("Failed to update Initialized to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
							return
						}
					}
				} else {
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerInitialized,
									Status:  core.ConditionFalse,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Initialized to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}

				// Todo: =========================== Check for Sealed status ==================================
				if hr.Sealed {
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerUnsealed,
									Status:  core.ConditionFalse,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Unsealed to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				} else {
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerUnsealed,
									Status:  core.ConditionTrue,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Unsealed to True for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerUnsealing,
									Status:  core.ConditionFalse,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Unsealing to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}

				if hr.Initialized && !hr.Sealed {
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerAcceptingConnection,
									Status:  core.ConditionTrue,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Accepting Conn. to True for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				} else {
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:    conapi.VaultServerAcceptingConnection,
									Status:  core.ConditionFalse,
									Message: "",
									Reason:  "",
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update Accepting Conn. to False for Vault Server: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}
			}
		}(vs)
	}
	// Wait until all go-routine complete executions
	wg.Wait()
}

//func (c *VaultController) getVaultClient(p *core.Pod) (*vaultapi.Client, error) {
//	// No need to use tunnel for StatefulSet
//	podAddr := util.PodDNSName(*p)
//	podPort := "8200"
//	tlsConfig := &vaultapi.TLSConfig{
//		Insecure: true,
//	}
//
//	vaultClient, err := util.NewVaultClient(podAddr, podPort, tlsConfig)
//	if err != nil {
//		return nil, errors.Wrapf(err, "failed creating client for the vault pod (%s/%s).", p.Namespace, p.Name)
//	}
//
//	return vaultClient, nil
//}

func (c *VaultController) getVaultServiceSpecificClient(vs *api.VaultServer) (*vaultapi.Client, error) {
	tlsConfig := &vaultapi.TLSConfig{
		Insecure: true,
	}

	svcURL := fmt.Sprintf("%s://%s.%s.svc:%d", vs.Scheme(), vs.ServiceName(api.VaultServerServiceVault), vs.Namespace, conapi.VaultAPIPort)

	vaultClient, err := util.NewVaultClient(svcURL, tlsConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating client for the vault service (%s/%s).", vs.Namespace, vs.ServiceName(api.VaultServerServiceVault))
	}

	return vaultClient, nil
}
