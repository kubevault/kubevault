package unsealer

import (
	"fmt"
	"path/filepath"
	"time"

	kutilcorev1 "github.com/appscode/kutil/core/v1"
	api "github.com/kubevault/operator/apis/core/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/unsealer/aws"
	"github.com/kubevault/operator/pkg/vault/unsealer/azure"
	"github.com/kubevault/operator/pkg/vault/unsealer/google"
	"github.com/kubevault/operator/pkg/vault/unsealer/kubernetes"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
)

type Unsealer interface {
	Apply(pt *corev1.PodTemplateSpec) error
	GetRBAC(namespace string) []rbac.Role
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
func (u *unsealerSrv) Apply(pt *corev1.PodTemplateSpec) error {
	if u == nil {
		return nil
	}

	var args []string
	vautlCACertFile := "/etc/vault/tls/ca.crt"
	cont := corev1.Container{
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

		pt.Spec.Volumes = kutilcorev1.UpsertVolume(pt.Spec.Volumes, corev1.Volume{
			Name: "vaultCA",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: u.VaultCASecret,
				},
			},
		})

		cont.VolumeMounts = kutilcorev1.UpsertVolumeMount(cont.VolumeMounts, corev1.VolumeMount{
			Name:      "vaultCA",
			MountPath: filepath.Dir(vautlCACertFile),
		})
	}

	cont.Args = append(cont.Args, args...)
	pt.Spec.Containers = kutilcorev1.UpsertContainer(pt.Spec.Containers, cont)
	err := u.unsealer.Apply(pt)
	if err != nil {
		return err
	}
	return nil
}

// GetRBAC return rbac roles required by unsealer
func (u *unsealerSrv) GetRBAC(namespace string) []rbac.Role {
	if u == nil {
		return nil
	}
	return u.unsealer.GetRBAC(namespace)
}
