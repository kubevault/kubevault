package controller

import (
	"strconv"
	"time"

	"github.com/appscode/go/log"
	v1u "github.com/appscode/kutil/core/v1"
	"github.com/hashicorp/vault/api"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type AppRole struct {
	BindSecretID    bool          `json:"bind_secret_id"`
	BoundCidrList   []string      `json:"bound_cidr_list"`
	Period          time.Duration `json:"period"`
	Policies        []string      `json:"policies"`
	SecretIDNumUses int           `json:"secret_id_num_uses"`
	SecretIDTTL     time.Duration `json:"secret_id_ttl"`
	TokenMaxTTL     time.Duration `json:"token_max_ttl"`
	TokenNumUses    int           `json:"token_num_uses"`
	TokenTTL        time.Duration `json:"token_ttl"`
}

func (c *VaultController) mountSecretBackend() error {
	// list enabled auth mechanism
	mounts, err := c.vaultClient.Sys().ListMounts()
	if err != nil {
		return err
	}
	for name, mnt := range mounts {
		if mnt.Type == "generic" && name == c.options.SecretBackend() {
			log.Infof("Found secret backend %s of type %s!", name, mnt.Type)
			return nil
		}
	}

	log.Infof("Enabling secret backend %s of type %s!", c.options.SecretBackend(), "generic")
	return c.vaultClient.Sys().Mount(c.options.SecretBackend(), &api.MountInput{
		Type: "generic",
	})
}

func (c *VaultController) mountAuthBackend() error {
	// list enabled auth mechanism
	mounts, err := c.vaultClient.Sys().ListAuth()
	if err != nil {
		return err
	}
	for name, mnt := range mounts {
		if mnt.Type == "approle" && name == c.options.AuthBackend() {
			log.Infof("Found auth backend %s of type %s!", name, mnt.Type)
			return nil
		}
	}

	log.Infof("Enabling auth backend %s of type %s!", c.options.AuthBackend(), "approle")
	return c.vaultClient.Sys().EnableAuthWithOptions(c.options.AuthBackend(), &api.EnableAuthOptions{
		Type: "approle",
	})
}

func (c *VaultController) renewTokens() {
	defer runtime.HandleCrash()
	for range c.renewer.C {
		list, err := c.k8sClient.CoreV1().Secrets(core.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			continue
		}
		for _, secret := range list.Items {
			if _, found := GetString(secret.Annotations, "kubernetes.io/service-account.name"); found {

				// secret.Data["LEASE_DURATION"]
				if renewable, _ := strconv.ParseBool(string(secret.Data["RENEWABLE"])); renewable {
					token := string(secret.Data[api.EnvVaultToken])
					vr, err := c.vaultClient.Auth().Token().RenewTokenAsSelf(token, 0)
					if err != nil {
						log.Errorln(err)
					}

					_, err = v1u.PatchSecret(c.k8sClient, &secret, func(in *core.Secret) *core.Secret {
						if in.Data == nil {
							in.Data = map[string][]byte{}
						}
						in.Data[api.EnvVaultToken] = []byte(vr.Auth.ClientToken)
						in.Data["VAULT_TOKEN_ACCESSOR"] = []byte(vr.Auth.Accessor)
						in.Data["LEASE_DURATION"] = []byte(strconv.Itoa(vr.Auth.LeaseDuration))
						in.Data["RENEWABLE"] = []byte(strconv.FormatBool(vr.Auth.Renewable))
						return in
					})
					if err != nil {
						log.Errorln(err)
					}
				}
			}
		}
	}
}
