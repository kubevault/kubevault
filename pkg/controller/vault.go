package controller

import (
	"time"

	"github.com/appscode/go/log"
	"github.com/hashicorp/vault/api"
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
