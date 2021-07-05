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

func (c *VaultController) RunHealthChecker(stopCh <-chan struct{}) {
	// As CheckVaultserverHealth() is a blocking function,
	// run it on a go-routine.
	go c.CheckVaultserverHealth(stopCh)
}

func (c *VaultController) CheckVaultserverHealth(stopCh <-chan struct{}) {
	klog.Info("Starting Elasticsearch health checker...")
	for {
		select {
		case <-stopCh:
			klog.Info("Shutting down Vaultserver health checker...")
			return
		default:
			c.CheckVaultserverHealthOnce()
			time.Sleep(10 * time.Second) // Todo: add as constant in apimachinery
		}
	}
}

func (c *VaultController) CheckVaultserverHealthOnce() {
	vsList, err := c.vsLister.VaultServers(core.NamespaceAll).List(labels.Everything())
	if err != nil {
		klog.Errorf("Failed to list Vaultserver objects with: %s", err.Error())
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

			// Todo: Algorithm:
			//- make a context with timeout (30 sec)
			//- make a list a pods, using the vs label selectors
			//
			//- for each pod:
			//	- make a vaultserver client (pod specific client)
			//	- health check call using the client
			//		- fail:
			//			- warning print log & continue
			//			- check error:
			//				- status code (network error):
			//					- do nothing
			//				- status code (anything else)
			//					- condition update (accepting connection - false)
			//
			//		- pass:
			//			- continue
			//
			//- if all health passed:
			//	- update condition to accepting connection true
			//- else:
			//	- failed

			// Todo: make a context with timeout (30 sec)
			// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			// defer cancel()

			// Todo: make a list a pods, using the vs label selectors
			name, namespace := vs.Name, vs.Namespace
			sel := vs.OffshootSelectors()

			opt := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(sel).String()}

			pods, err := c.kubeClient.CoreV1().Pods(namespace).List(context.TODO(), opt)
			if err != nil {
				klog.Errorf("=== failed listing pods for the vault server (%s.%s): %v ===", namespace, name, err)
				return
			}

			if len(pods.Items) == 0 {
				klog.Errorf("=== for the vault server (%s.%s): no pods found ===", namespace, name)
				return
			}

			failed := false
			// Todo: iterate over each pod
			for _, pod := range pods.Items {
				// Todo: make a pod specific vaultserver client
				vaultClient, err := c.getVaultClient(&pod)
				if err != nil {
					klog.Errorf("=== failed creating client for the vault pod (%s/%s). ===", pod.Namespace, pod.Name)
					continue
				}

				// Todo:  make the health check call using the vaultClient
				hr, err := vaultClient.Sys().Health()
				if err != nil {
					failed = true
					klog.Warningf(" === failed requesting health info for the vault pod (%s/%s). ===", pod.Namespace, pod.Name)

					//200 if initialized, unsealed, and active
					//429 if unsealed and standby
					//472 if disaster recovery mode replication secondary and active
					//473 if performance standby
					//501 if not initialized
					//503 if sealed

					// if hr is nil, it will not check the other conditions
					if hr == nil || !hr.Initialized || hr.Sealed {
						continue
					}
					klog.Infoln("=========== HR status ==============", hr.Sealed, hr.Initialized)
					_, err = cs_util.UpdateVaultServerStatus(
						context.TODO(),
						c.extClient.KubevaultV1alpha1(),
						vs.ObjectMeta,
						func(in *api.VaultServerStatus) *api.VaultServerStatus {
							in.Conditions = kmapi.SetCondition(in.Conditions,
								kmapi.Condition{
									Type:   conapi.VaultserverAcceptingConnection,
									Status: core.ConditionFalse,
								})
							return in
						},
						metav1.UpdateOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to update status for Vaultserver: %s/%s with %s", vs.Namespace, vs.Name, err.Error())
						return
					}
				}
			}
			if !failed {
				klog.Infoln("========================= NOT Failed =========================")
				_, err = cs_util.UpdateVaultServerStatus(
					context.TODO(),
					c.extClient.KubevaultV1alpha1(),
					vs.ObjectMeta,
					func(in *api.VaultServerStatus) *api.VaultServerStatus {
						in.Conditions = kmapi.SetCondition(in.Conditions,
							kmapi.Condition{
								Type:   conapi.VaultserverAcceptingConnection,
								Status: core.ConditionTrue,
							})
						return in
					},
					metav1.UpdateOptions{},
				)
			} else {
				klog.Infoln("========================= Some POD Failed =========================")
			}
		}(vs)
	}

	// Wait until all go-routine complete executions
	wg.Wait()
}

func (c *VaultController) getVaultClient(p *core.Pod) (*vaultapi.Client, error) {
	// No need to use tunnel for StatefulSet
	podAddr := util.PodDNSName(*p)
	podPort := "8200"
	tlsConfig := &vaultapi.TLSConfig{
		Insecure: true,
	}

	vaultClient, err := util.NewVaultClient(podAddr, podPort, tlsConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating client for the vault pod (%s/%s).", p.Namespace, p.Name)
	}

	return vaultClient, nil
}
