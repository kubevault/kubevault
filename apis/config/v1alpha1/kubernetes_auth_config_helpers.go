package v1alpha1

func (c *KubernetesAuthConfiguration) SetDefaults() {
	if c == nil {
		return
	}

	if c.AuthPath == "" {
		c.AuthPath = "kubernetes"
	}
}
