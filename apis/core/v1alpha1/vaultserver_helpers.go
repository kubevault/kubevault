package v1alpha1

func (v VaultServer) OffshootName() string {
	return v.Name
}

func (v VaultServer) OffshootLabels() map[string]string {
	return map[string]string{
		"app":           "vault",
		"vault_cluster": v.Name,
	}
}

func (v VaultServer) ServiceName() string {
	return v.OffshootName()
}

func (v VaultServer) DeploymentName() string {
	return v.OffshootName()
}

func (v VaultServer) ServiceAccountName() string {
	return v.OffshootName()
}

func (v VaultServer) ConfigMapName() string {
	return v.OffshootName() + "-vault-config"
}

func (v VaultServer) TLSSecretName() string {
	return v.OffshootName() + "-vault-tls"
}

// SetDefaults sets the default values for the vault spec and returns true if the spec was changed
func (v *VaultServer) SetDefaults() bool {
	changed := false
	vs := &v.Spec
	if vs.Nodes == 0 {
		vs.Nodes = 1
		changed = true
	}
	if len(vs.BaseImage) == 0 {
		vs.BaseImage = defaultBaseImage
		changed = true
	}
	if len(vs.Version) == 0 {
		vs.Version = defaultVersion
		changed = true
	}
	return changed
}
