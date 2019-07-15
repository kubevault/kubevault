package s3

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
)

func TestOptions_GetStorageConfig(t *testing.T) {
	opts, err := NewOptions(api.S3Spec{
		Bucket:             "test",
		EndPoint:           "endpoint",
		Region:             "test-region",
		CredentialSecret:   "credential",
		SessionTokenSecret: "session",
		MaxParallel:        128,
		S3ForcePathStyle:   true,
		DisableSSL:         true,
	})
	assert.Nil(t, err)

	out := `
storage "s3" {
bucket = "test"
endpoint = "endpoint"
region = "test-region"
s3_force_path_style = "true"
disable_ssl = "true"
max_parallel = 128
}
`
	t.Run("S3 storage config", func(t *testing.T) {
		got, err := opts.GetStorageConfig()
		assert.Nil(t, err)
		if !assert.Equal(t, out, got) {
			fmt.Println("expected:", out)
			fmt.Println("got:", got)
		}
	})
}
