package v1alpha1

func (c *CertAuthConfiguration) SetDefaults() {
	if c == nil {
		return
	}

	if c.AuthPath == "" {
		c.AuthPath = "cert"
	}
}
