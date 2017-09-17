package controller

import (
	"github.com/appscode/go/log"
	"github.com/hashicorp/vault/api"
)

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
