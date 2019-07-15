package swift

import (
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
)

var swiftStorageFmt = `
storage "swift" {
%s
}
`

type Options struct {
	api.SwiftSpec
}

func NewOptions(s api.SwiftSpec) (*Options, error) {
	return &Options{
		s,
	}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	if o.CredentialSecret != "" {
		var envs []core.EnvVar
		envs = append(envs, core.EnvVar{
			Name: "OS_USERNAME",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "username",
				},
			},
		}, core.EnvVar{
			Name: "OS_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: o.CredentialSecret,
					},
					Key: "password",
				},
			},
		})

		if o.AuthTokenSecret != "" {
			envs = append(envs, core.EnvVar{
				Name: "OS_AUTH_TOKEN",
				ValueFrom: &core.EnvVarSource{
					SecretKeyRef: &core.SecretKeySelector{
						LocalObjectReference: core.LocalObjectReference{
							Name: o.AuthTokenSecret,
						},
						Key: "auth_token",
					},
				},
			})
		}

		pt.Spec.Containers[0].Env = append(pt.Spec.Containers[0].Env, envs...)
	}
	return nil
}

// vault doc: https://www.vaultproject.io/docs/configuration/storage/swift.html
//
//  GetStorageConfig creates swift storage config from SwiftSpec
func (o *Options) GetStorageConfig() (string, error) {
	params := []string{}
	if o.AuthUrl != "" {
		params = append(params, fmt.Sprintf(`auth_url = "%s"`, o.AuthUrl))
	}
	if o.Container != "" {
		params = append(params, fmt.Sprintf(`container = "%s"`, o.Container))
	}
	if o.Tenant != "" {
		params = append(params, fmt.Sprintf(`tenant = "%s"`, o.Tenant))
	}
	if o.MaxParallel != 0 {
		params = append(params, fmt.Sprintf(`max_parallel = "%d"`, o.MaxParallel))
	}
	if o.Region != "" {
		params = append(params, fmt.Sprintf(`region = "%s"`, o.Region))
	}
	if o.TenantID != "" {
		params = append(params, fmt.Sprintf(`tenant_id = "%s"`, o.TenantID))
	}
	if o.Domain != "" {
		params = append(params, fmt.Sprintf(`domain = "%s"`, o.Domain))
	}
	if o.ProjectDomain != "" {
		params = append(params, fmt.Sprintf(`project-domain = "%s"`, o.ProjectDomain))
	}
	if o.TrustID != "" {
		params = append(params, fmt.Sprintf(`trust_id = "%s"`, o.TrustID))
	}
	if o.StorageUrl != "" {
		params = append(params, fmt.Sprintf(`storage_url = "%s"`, o.StorageUrl))
	}

	storageCfg := fmt.Sprintf(swiftStorageFmt, strings.Join(params, "\n"))
	return storageCfg, nil
}
