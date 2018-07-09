package dynamodb

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var dynamodbStorageFmt = `
storage "dynamodb" {
%s
}
`

type Options struct {
	api.DynamoDBSpec
}

func NewOptions(s api.DynamoDBSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	envs := []corev1.EnvVar{}

	if o.CredentialSecret != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "access_key",
				},
			},
		}, corev1.EnvVar{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "secret_key",
				},
			},
		})
	}
	if o.SessionTokenSecret != "" {
		envs = append(envs, corev1.EnvVar{
			Name: "AWS_SESSION_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
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
// GetStorageConfig creates gcs storae config from GcsSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.EndPoint != "" {
		params = append(params, fmt.Sprintf(`endpoint = "%s"`, o.EndPoint))
	}
	if o.HaEnabled == true {
		params = append(params, fmt.Sprintf(`ha_enable = "true"`))
	}
	if o.Region != "" {
		params = append(params, fmt.Sprintf(`region = "%s"`, o.Region))
	}
	if o.ReadCapacity != 0 {
		params = append(params, fmt.Sprintf(`read_capacity = %d`, o.ReadCapacity))
	}
	if o.WriteCapacity != 0 {
		params = append(params, fmt.Sprintf(`write_capacity = %d`, o.WriteCapacity))
	}
	if o.Table != "" {
		params = append(params, fmt.Sprintf(`table = "%s"`, o.Table))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = %d`, o.MaxParallel))
	}

	storageCfg := fmt.Sprintf(dynamodbStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
