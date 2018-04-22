package controller

import (
	"fmt"
	"path"
	"time"

	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (c *VaultController) initSecretWatcher() {
	c.sInformer = c.kubeInformerFactory.InformerFor(&core.Secret{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return core_informers.NewFilteredSecretInformer(
			client,
			core.NamespaceAll,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			nil,
		)
	})
	c.sQueue = queue.New("Secret", c.MaxNumRequeues, c.NumThreads, c.syncSecretToVault)
	c.sInformer.AddEventHandler(queue.DefaultEventHandler(c.sQueue.GetQueue()))
	// c.sLister = c.informerFactory.Apps().V1beta1().StatefulSets().Lister()
}

func (c *VaultController) syncSecretToVault(key string) error {
	obj, exists, err := c.sInformer.GetIndexer().GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Secret, so that we will see a delete for one secret
		fmt.Printf("Secret %s does not exist anymore\n", key)

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}
		c.vaultClient.Logical().Delete(path.Join(c.SecretBackend(), namespace, name))
	} else {
		secret := obj.(*core.Secret)
		fmt.Printf("Sync/Add/Update for Secret %s\n", secret.GetName())

		if secret.DeletionTimestamp != nil {
			if core_util.HasFinalizer(secret.ObjectMeta, "finalizer.kubernetes.io/vault") {
				saName, saNameFound := GetString(secret.Annotations, "kubernetes.io/service-account.name")
				saUID, saUIDFound := GetString(secret.Annotations, "kubernetes.io/service-account.uid")
				if saNameFound && saUIDFound {
					err := c.vaultClient.Auth().Token().RevokeAccessor(string(secret.Data["VAULT_TOKEN_ACCESSOR"]))
					if err != nil {
						log.Errorln(err)
					}
					log.Infof("Revoked token accessor %s", string(secret.Data["VAULT_TOKEN_ACCESSOR"]))

					// create new secret for rolename if s/a still exists
					sa, err := c.kubeClient.CoreV1().ServiceAccounts(secret.Namespace).Get(saName, metav1.GetOptions{})
					if err == nil && string(sa.UID) == saUID {
						newSecret, err := c.createVaultToken(sa)
						if err != nil {
							return err
						}

						_, _, err = core_util.CreateOrPatchServiceAccount(c.kubeClient, metav1.ObjectMeta{Namespace: sa.Namespace, Name: sa.Name}, func(in *core.ServiceAccount) *core.ServiceAccount {
							if in.Annotations == nil {
								in.Annotations = map[string]string{}
							}
							in.Annotations["vaultproject.io/secret.name"] = newSecret.Name
							return in
						})
						if err != nil {
							return err
						}
					} else {
						log.Errorln(err)
					}
				}
				core_util.PatchSecret(c.kubeClient, secret, func(in *core.Secret) *core.Secret {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, "finalizer.kubernetes.io/vault")
					return in
				})

			}
			return nil
		}

		if _, found := GetString(secret.Annotations, "vaultproject.io/origin"); !found {
			data := map[string]interface{}{}
			for k, v := range secret.Data {
				data[k] = string(v)
			}
			data["m:uid"] = secret.ObjectMeta.UID
			data["m:resourceVersion"] = secret.ObjectMeta.ResourceVersion
			data["m:generation"] = secret.ObjectMeta.Generation
			data["m:creationTimestamp"] = secret.ObjectMeta.CreationTimestamp.UTC().Format(time.RFC3339)
			if secret.ObjectMeta.DeletionTimestamp != nil {
				data["m:deletionTimestamp"] = secret.ObjectMeta.DeletionTimestamp.UTC().Format(time.RFC3339)
			}
			for k, v := range secret.ObjectMeta.Labels {
				data["l:"+k] = v
			}
			for k, v := range secret.ObjectMeta.Annotations {
				data["a:"+k] = v
			}
			_, err = c.vaultClient.Logical().Write(path.Join(c.SecretBackend(), secret.Namespace, secret.Name), data)
			return err
		} else {
			_, _, err = core_util.PatchSecret(c.kubeClient, secret, func(in *core.Secret) *core.Secret {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, "finalizer.kubernetes.io/vault")
				return in
			})
			return err
		}
	}
	return nil
}
