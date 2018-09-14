package aws

import (
	"fmt"

	kutilcorev1 "github.com/appscode/kutil/core/v1"
	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
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

func (o *Options) Apply(pt *corev1.PodTemplateSpec) error {
	if pt == nil {
		return errors.New("podTempleSpec is nil")
	}

	var args []string
	var cont corev1.Container

	for _, c := range pt.Spec.Containers {
		if c.Name == util.VaultUnsealerContainerName() {
			cont = c
		}
	}

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

	cont.Env = kutilcorev1.UpsertEnvVars(cont.Env, envs...)
	pt.Spec.Containers = kutilcorev1.UpsertContainer(pt.Spec.Containers, cont)
	return nil
}

// GetRBAC returns required rbac roles
func (o *Options) GetRBAC(namespace string) []rbac.Role {
	return nil
}
