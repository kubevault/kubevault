package s3

import (
	"fmt"
	"strings"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	core "k8s.io/api/core/v1"
)

var s3StorageFmt = `
storage "s3" {
%s
}
`

type Options struct {
	api.S3Spec
}

func NewOptions(s api.S3Spec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

// Set environment variable:
//	- AWS_ACCESS_KEY_ID
//	- AWS_SECRET_ACCESS_KEY
//  - AWS_SESSION_TOKEN
func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	envs := []core.EnvVar{}
	if o.CredentialSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "access_key",
				},
			},
		}, core.EnvVar{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "secret_key",
				},
			},
		})
	}
	if o.SessionTokenSecret != "" {
		envs = append(envs, core.EnvVar{
			Name: "AWS_SESSION_TOKEN",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.SessionTokenSecret,
					},
					Key: "session_token",
				},
			},
		})
	}

	pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, envs...)

	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/google-cloud-storage.html
//
//  GetStorageConfig creates gcs storae config from S3Spec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.Bucket != "" {
		params = append(params, fmt.Sprintf(`bucket = "%s"`, o.Bucket))
	}
	if o.EndPoint != "" {
		params = append(params, fmt.Sprintf(`endpoint = "%s"`, o.EndPoint))
	}
	if o.Region != "" {
		params = append(params, fmt.Sprintf(`region = "%s"`, o.Region))
	}
	if o.S3ForcePathStyle {
		params = append(params, fmt.Sprintf(`s3_force_path_style = "true"`))
	}
	if o.DisableSSL {
		params = append(params, fmt.Sprintf(`disable_ssl = "true"`))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = %d`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(s3StorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
