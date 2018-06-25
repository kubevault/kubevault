package aws

import (
	"fmt"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
)

const (
	ModeAwsKmsSsm = "aws-kms-ssm"
)

type Options struct {
	api.AwsKmsSsmSpec
}

func NewOptions(s api.AwsKmsSsmSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *corev1.PodTemplateSpec, cont *corev1.Container) error {
	var args []string

	args = append(args, fmt.Sprintf("--mode=%s", ModeAwsKmsSsm))

	if o.KmsKeyID != "" {
		args = append(args, fmt.Sprintf("--aws.kms-key-id=%s", o.KmsKeyID))
	}

	cont.Args = append(cont.Args, args...)

	var envs []corev1.EnvVar
	if o.Region != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "AWS_REGION",
			Value: o.Region,
		})
	}
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

	cont.Env = append(cont.Env, envs...)

	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(namespace string) []rbac.Role {
	return nil
}

func (o *Options) GetSecrets(namespace string) ([]corev1.Secret, error) {
	return nil, nil
}
