package swift

import (
	"fmt"
	"testing"

	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestOptions_GetStorageConfig(t *testing.T) {
	opts, err := NewOptions(api.SwiftSpec{
		AuthUrl:          "auth",
		Container:        "vault",
		Tenant:           "tenant",
		CredentialSecret: "cred",
		MaxParallel:      128,
		Region:           "bn",
		TenantID:         "1234",
		Domain:           "hi.com",
		ProjectDomain:    "p.com",
		TrustID:          "1234",
		StorageUrl:       "s.com",
		AuthTokenSecret:  "auth",
	})
	assert.Nil(t, err)

	out := `
storage "swift" {
auth_url = "auth"
container = "vault"
tenant = "tenant"
max_parallel = "128"
region = "bn"
tenant_id = "1234"
domain = "hi.com"
project-domain = "p.com"
trust_id = "1234"
storage_url = "s.com"
}
`
	t.Run("Swift storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)
		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
