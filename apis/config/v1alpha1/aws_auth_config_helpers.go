package v1alpha1

func (c *AWSAuthConfiguration) SetDefaults() {
	if c == nil {
		return
	}

	if c.AuthPath == "" {
		c.AuthPath = "aws"
	}
}
