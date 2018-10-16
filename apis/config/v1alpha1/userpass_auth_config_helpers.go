package v1alpha1

func (c *UserPassAuthConfiguration) SetDefaults() {
	if c == nil {
		return
	}

	if c.AuthPath == "" {
		c.AuthPath = "userpass"
	}
}
