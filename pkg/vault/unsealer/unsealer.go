package unsealer

import (
	"fmt"
	"path/filepath"
	"time"

	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/unsealer/aws"
	"github.com/kubevault/operator/pkg/vault/unsealer/azure"
	"github.com/kubevault/operator/pkg/vault/unsealer/google"
	"github.com/kubevault/operator/pkg/vault/unsealer/kubernetes"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
)

type Unsealer interface {
	Apply(pt *core.PodTemplateSpec) error
	GetRBAC(prefix, namespace string) []rbac.Role
}

type unsealerSrv struct {
	*api.UnsealerSpec
	unsealer Unsealer
	image    string
}

func newUnsealer(s *api.UnsealerSpec) (Unsealer, error) {
	if s.Mode.KubernetesSecret != nil {
		return kubernetes.NewOptions(*s.Mode.KubernetesSecret)
	} else if s.Mode.GoogleKmsGcs != nil {
		return google.NewOptions(*s.Mode.GoogleKmsGcs)
	} else if s.Mode.AwsKmsSsm != nil {
		return aws.NewOptions(*s.Mode.AwsKmsSsm)
	} else if s.Mode.AzureKeyVault != nil {
		return azure.NewOptions(*s.Mode.AzureKeyVault)
	} else {
		return nil, errors.New("unsealer mode is not valid/defined")
	}
}

func NewUnsealerService(s *api.UnsealerSpec, image string) (Unsealer, error) {
	if s == nil {
		return nil, nil
	}

	unslr, err := newUnsealer(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unsealer service")
	}
	return &unsealerSrv{
		UnsealerSpec: s,
		image:        image,
		unsealer:     unslr,
	}, nil
}

// Apply will do:
// 	- add unsealer container for vault
//	- add additional env, volume mounts etc if required
func (u *unsealerSrv) Apply(pt *core.PodTemplateSpec) error {
	if u == nil {
		return nil
	}

	var args []string
	vautlCACertFile := "/etc/vault/tls/ca.crt"
	cont := core.Container{
		Name:  util.VaultUnsealerContainerName,
		Image: u.image,
	}
	args = append(args,
		"run",
		"--v=3",
	)

	if u.SecretShares != 0 {
		args = append(args, fmt.Sprintf("--secret-shares=%d", u.SecretShares))
	}
	if u.SecretThreshold != 0 {
		args = append(args, fmt.Sprintf("--secret-threshold=%d", u.SecretThreshold))
	}

	if u.RetryPeriodSeconds != 0 {
		p := time.Second * u.RetryPeriodSeconds
		args = append(args, fmt.Sprintf("--retry-period=%s", p.String()))
	}
	if u.InsecureTLS == true {
		args = append(args, fmt.Sprintf("--insecure-tls=true"))
	}
	if u.OverwriteExisting == true {
		args = append(args, fmt.Sprintf("--overwrite-existing=true"))
	}

	if u.InsecureTLS == false && u.VaultCASecret != "" {
		args = append(args, fmt.Sprintf("--ca-cert-file=%s", vautlCACertFile))

		pt.Spec.Volumes = core_util.UpsertVolume(pt.Spec.Volumes, core.Volume{
			Name: "vaultCA",
			VolumeSource: core.VolumeSource{
				Secret: &core.SecretVolumeSource{
					SecretName: u.VaultCASecret,
				},
			},
		})

		cont.VolumeMounts = core_util.UpsertVolumeMount(cont.VolumeMounts, core.VolumeMount{
			Name:      "vaultCA",
			MountPath: filepath.Dir(vautlCACertFile),
		})
	}

	cont.Args = append(cont.Args, args...)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	err := u.unsealer.Apply(pt)
	if err != nil {
		return err
	}
	return nil
}

// GetRBAC return rbac roles required by unsealer
func (u *unsealerSrv) GetRBAC(prefix, namespace string) []rbac.Role {
	if u == nil {
		return nil
	}
	return u.unsealer.GetRBAC(prefix, namespace)
}
