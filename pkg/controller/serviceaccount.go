package controller

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/appscode/go/crypto/rand"
	v1_util "github.com/appscode/kutil/core/v1"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

func (c *VaultController) runServiceAccountWatcher() {
	for c.processNextServiceAccount() {
	}
}

func (c *VaultController) processNextServiceAccount() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.saQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two serviceAccounts with the same key are never processed in
	// parallel.
	defer c.saQueue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncServiceAccountToVault(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.saQueue.Forget(key)
		return true
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.saQueue.NumRequeues(key) < c.options.MaxNumRequeues {
		glog.Infof("Error syncing serviceAccount %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.saQueue.AddRateLimited(key)
		return true
	}

	c.saQueue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping serviceAccount %q out of the queue: %v", key, err)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the serviceAccount to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *VaultController) syncServiceAccountToVault(key string) error {
	obj, exists, err := c.saIndexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a ServiceAccount, so that we will see a delete for one serviceAccount
		fmt.Printf("ServiceAccount %s does not exist anymore\n", key)

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}
		roleName := namespace + "." + name
		p := path.Join("auth", c.options.AuthBackend(), "role", roleName)
		_, err = c.vaultClient.Logical().Delete(p)
		if err != nil {
			return err
		}
	} else {
		sa := obj.(*v1.ServiceAccount)
		fmt.Printf("Sync/Add/Update for ServiceAccount %s\n", sa.GetName())
		roleName := sa.Namespace + "." + sa.Name
		p := path.Join("auth", c.options.AuthBackend(), "role", roleName)

		resp, err := c.vaultClient.Logical().Read(p)
		if err != nil {
			return err
		}
		if resp == nil {
			_, err := c.vaultClient.Logical().Write(p, map[string]interface{}{
				"secret_id_num_uses": 1,
				"secret_id_ttl":      60,
				"token_num_uses":     0,
				"token_ttl":          3600,
				"token_max_ttl":      0,
				"period":             3600,
			})
			if err != nil {
				return err
			}
		}

		secretName, annotated := GetString(sa.Annotations, "vaultproject.io/secret.name")
		if !annotated {
			// auto generate name
			secretName = sa.Name + "-vault-" + rand.Characters(5)
		}
		var caCert []byte
		if c.options.CACertFile != "" {
			caCert, err = ioutil.ReadFile(c.options.CACertFile)
			if err != nil {
				return err
			}
		}
		_, err = v1_util.CreateOrPatchSecret(c.k8sClient, metav1.ObjectMeta{Namespace: sa.Namespace, Name: secretName}, func(in *v1.Secret) *v1.Secret {
			if in.Annotations == nil {
				in.Annotations = map[string]string{}
			}
			in.Annotations["kubernetes.io/service-account.name"] = sa.Name
			in.Annotations["kubernetes.io/service-account.uid"] = string(sa.UID)
			in.Annotations["vaultproject.io/approle.name"] = roleName

			if in.Labels == nil {
				in.Labels = map[string]string{}
			}
			in.Labels["kubernetes.io/vault-token"] = ""

			if in.Data == nil {
				in.Data = map[string][]byte{}
			}
			in.Data[api.EnvVaultAddress] = []byte(c.options.VaultAddress)
			if caCert != nil {
				in.Data[api.EnvVaultCACert] = caCert
			}

			fi := -1
			for i, ref := range in.OwnerReferences {
				if ref.Kind == "ServiceAccount" && ref.Name == sa.Name {
					fi = i
					break
				}
			}
			if fi == -1 {
				in.OwnerReferences = append(in.OwnerReferences, metav1.OwnerReference{})
				fi = len(in.OwnerReferences) - 1
			}
			in.OwnerReferences[fi].APIVersion = v1.SchemeGroupVersion.String()
			in.OwnerReferences[fi].Kind = "ServiceAccount"
			in.OwnerReferences[fi].Name = sa.Name
			in.OwnerReferences[fi].UID = sa.UID

			return in
		})
		if err != nil {
			return err
		}

		if !annotated {
			_, err = v1_util.CreateOrPatchServiceAccount(c.k8sClient, metav1.ObjectMeta{Namespace: sa.Namespace, Name: sa.Name}, func(in *v1.ServiceAccount) *v1.ServiceAccount {
				if in.Annotations == nil {
					in.Annotations = map[string]string{}
				}
				in.Annotations["vaultproject.io/secret.name"] = secretName
				return in
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetString(m map[string]string, k string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, found := m[k]
	return v, found
}
