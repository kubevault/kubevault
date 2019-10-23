package azure

import (
	"fmt"
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestOptions_GetStorageConfig(t *testing.T) {
	opts, err := NewOptions(api.AzureSpec{
		AccountName: "ac",
		Container:   "vault",
		MaxParallel: 111,
	})
	assert.Nil(t, err)

	out := `
storage "azure" {
accountName = "ac"
container = "vault"
max_parallel = 111
}
`
	t.Run("Azure storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)
		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
