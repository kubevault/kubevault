package postgresql

import (
	"fmt"
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestOptions_GetStorageConfig(t *testing.T) {
	opts, err := NewOptions(api.PostgreSQLSpec{
		ConnectionUrl: "url",
		Table:         "table",
		MaxParallel:   128,
	})
	assert.Nil(t, err)

	out := `
storage "postgresql" {
connection_url = "url"
table = "table"
max_parallel = "128"
}
`
	t.Run("PostgreSQL storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)
		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
