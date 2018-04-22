package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"github.com/appscode/go/crypto/rand"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/golang/glog"
	"github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (c *VaultController) initServiceAccountWatcher() {
	c.saInformer = c.kubeInformerFactory.InformerFor(&core.ServiceAccount{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return core_informers.NewFilteredServiceAccountInformer(
			client,
			core.NamespaceAll,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			nil,
		)
	})
	c.saQueue = queue.New("ServiceAccount", c.MaxNumRequeues, c.NumThreads, c.syncServiceAccountToVault)
	c.saInformer.AddEventHandler(queue.DefaultEventHandler(c.saQueue.GetQueue()))
	// c.saLister = c.informerFactory.Apps().V1beta1().StatefulSets().Lister()
}

func (c *VaultController) syncServiceAccountToVault(key string) error {
	obj, exists, err := c.saInformer.GetIndexer().GetByKey(key)
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
		p := path.Join("auth", c.AuthBackend(), "role", roleName)
		_, err = c.vaultClient.Logical().Delete(p)
		if err != nil {
			return err
		}
	} else {
		sa := obj.(*core.ServiceAccount)
		fmt.Printf("Sync/Add/Update for ServiceAccount %s\n", sa.GetName())

		p := path.Join("auth", c.AuthBackend(), "role", c.roleName(sa))
		resp, err := c.vaultClient.Logical().Read(p)
		if err != nil {
			return err
		}
		if resp == nil {
			_, err := c.vaultClient.Logical().Write(p, map[string]interface{}{
				"secret_id_num_uses": 1,
				"secret_id_ttl":      60,
				"token_num_uses":     0,
				"token_ttl":          "168h",
				"token_max_ttl":      0,
				"period":             "168h",
			})
			if err != nil {
				return err
			}
		} else {
			if maxTTL := GetInt(resp.Data, "token_max_ttl"); maxTTL != 0 {
				_, err := c.vaultClient.Logical().Write(path.Join(p, "token-max-ttl"), map[string]interface{}{
					"token_max_ttl": 0,
				})
				if err != nil {
					return err
				}
			}
			if ttl := GetInt(resp.Data, "token_ttl"); ttl == 0 {
				_, err := c.vaultClient.Logical().Write(path.Join(p, "token-ttl"), map[string]interface{}{
					"token_ttl": "168h",
				})
				if err != nil {
					return err
				}
			}
			if period := GetInt(resp.Data, "period"); period == 0 {
				_, err := c.vaultClient.Logical().Write(path.Join(p, "period"), map[string]interface{}{
					"period": "168h",
				})
				if err != nil {
					return err
				}
			}
		}

		secretName, annotated := GetString(sa.Annotations, "vaultproject.io/secret.name")
		if annotated {
			_, err = c.kubeClient.CoreV1().Secrets(sa.Namespace).Get(secretName, metav1.GetOptions{})
			if kerr.IsNotFound(err) {
				annotated = false
			}
		}
		if !annotated {
			secret, err := c.createVaultToken(sa)
			if err != nil {
				return err
			}

			_, _, err = core_util.CreateOrPatchServiceAccount(c.kubeClient, metav1.ObjectMeta{Namespace: sa.Namespace, Name: sa.Name}, func(in *core.ServiceAccount) *core.ServiceAccount {
				if in.Annotations == nil {
					in.Annotations = map[string]string{}
				}
				in.Annotations["vaultproject.io/secret.name"] = secret.Name
				return in
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *VaultController) roleName(sa *core.ServiceAccount) string {
	return sa.Namespace + "." + sa.Name
}

func (c *VaultController) createVaultToken(sa *core.ServiceAccount) (*core.Secret, error) {
	var caCert []byte
	var err error
	if c.CACertFile != "" {
		caCert, err = ioutil.ReadFile(c.CACertFile)
		if err != nil {
			return nil, err
		}
	}

	vr, err := c.vaultClient.Logical().Read(path.Join("auth", c.AuthBackend(), "role", c.roleName(sa), "role-id"))
	if err != nil {
		return nil, err
	}
	roleID := vr.Data["role_id"]

	mdSecret, err := json.Marshal(map[string]string{
		"kind":      "ServiceAccount",
		"name":      sa.Name,
		"namespace": sa.Namespace,
		"uid":       string(sa.UID),
	})
	if err != nil {
		return nil, err
	}
	vr, err = c.vaultClient.Logical().Write(path.Join("auth", c.AuthBackend(), "role", c.roleName(sa), "secret-id"), map[string]interface{}{
		"metadata": string(mdSecret),
	})
	if err != nil {
		return nil, err
	}
	secretID := vr.Data["secret_id"]

	vr, err = c.vaultClient.Logical().Write(path.Join("auth", c.AuthBackend(), "login"), map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	})
	if err != nil {
		return nil, err
	}
	if vr.Auth == nil {
		return nil, fmt.Errorf("missing token for role %s", c.roleName(sa))
	}

	_, err = c.vaultClient.Logical().Write(path.Join("auth", c.AuthBackend(), "role", c.roleName(sa), "secret-id", "destroy"), map[string]interface{}{
		"secret_id": secretID,
	})
	if err != nil {
		return nil, err
	}

	// auto generate name
	secretName := sa.Name + "-vault-" + rand.Characters(5)
	out, _, err := core_util.CreateOrPatchSecret(c.kubeClient, metav1.ObjectMeta{Namespace: sa.Namespace, Name: secretName}, func(in *core.Secret) *core.Secret {
		if in.Annotations == nil {
			in.Annotations = map[string]string{}
		}
		in.Annotations["kubernetes.io/service-account.name"] = sa.Name
		in.Annotations["kubernetes.io/service-account.uid"] = string(sa.UID)
		in.Annotations["vaultproject.io/approle.name"] = c.roleName(sa)
		in.Annotations["vaultproject.io/origin"] = "steward"

		if in.Labels == nil {
			in.Labels = map[string]string{}
		}
		in.Labels["kubernetes.io/vault-token"] = ""

		if in.Data == nil {
			in.Data = map[string][]byte{}
		}
		in.Data[api.EnvVaultAddress] = []byte(c.VaultAddress)
		if caCert != nil {
			in.Data[api.EnvVaultCACert] = caCert
		}
		in.Data[api.EnvVaultToken] = []byte(vr.Auth.ClientToken)
		in.Data["VAULT_TOKEN_ACCESSOR"] = []byte(vr.Auth.Accessor)
		in.Data["LEASE_DURATION"] = []byte(strconv.Itoa(vr.Auth.LeaseDuration))
		in.Data["RENEWABLE"] = []byte(strconv.FormatBool(vr.Auth.Renewable))

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
		in.OwnerReferences[fi].APIVersion = core.SchemeGroupVersion.String()
		in.OwnerReferences[fi].Kind = "ServiceAccount"
		in.OwnerReferences[fi].Name = sa.Name
		in.OwnerReferences[fi].UID = sa.UID

		return in
	})
	return out, err
}

func GetString(m map[string]string, k string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, found := m[k]
	return v, found
}

func GetInt(m map[string]interface{}, k string) int {
	if m == nil {
		return 0
	}
	v, found := m[k]
	if !found {
		return 0
	}
	switch u := v.(type) {
	case json.Number:
		i, _ := u.Int64()
		return int(i)
	case int:
		return u
	case int64:
		return int(u)
	case int32:
		return int(u)
	case int16:
		return int(u)
	case int8:
		return int(u)
	}
	return 0
}
