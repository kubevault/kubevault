package unsealer

import (
	"fmt"
	"path/filepath"
	"time"

	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	"github.com/kube-vault/operator/pkg/vault/unsealer/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
)

type UnsealerService interface {
	Apply(pt *corev1.Container) error
	GetRBAC(namespace string, lable map[string]string) []rbac.Role
}

type Unsealer struct {
	Service UnsealerService
	*api.UnsealerSpec
}

func NewUnsealerService(s *api.UnsealerSpec) (UnsealerService, error) {
	if s.Mode.KubernetesSecret != nil {
		return kubernetes.NewOptions(*s.Mode.KubernetesSecret)
	} else {
		return nil, errors.New("unsealer mode is not valid/defined")
	}
}

func NewUnsealer(s *api.UnsealerSpec) (*Unsealer, error) {
	srv, err := NewUnsealerService(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unsealer service")
	}

	return &Unsealer{
		srv,
		s,
	}, nil
}

// AddContainer add unsealer container for vault
func (u *Unsealer) AddContainer(pt *corev1.PodTemplateSpec) error {
	var args []string

	vautlCACertFile := "/etc/vault/tls/ca.crt"

	cont := corev1.Container{
		Name:  "vault-unsealer",
		Image: "nightfury1204/vault-unsealer:canary",
	}

	args = append(args, "run")

	if u.SecretShares != 0 {
		args = append(args, fmt.Sprintf("--secret-shares=%d", u.SecretShares))
	}
	if u.SecretThreshold != 0 {
		args = append(args, fmt.Sprintf("--secret-threshold=%d", u.SecretThreshold))
	}

	// TODO: keep this?
	/*if u.StoreRootToken == false {
		args = append(args, fmt.Sprintf("--store-root-token=false"))
	} else {
		args = append(args, fmt.Sprintf("--store-root-token=true"))
	}*/

	if u.RetryPeriodSeconds != 0 {
		p := time.Second * u.RetryPeriodSeconds
		args = append(args, fmt.Sprintf("--retry-period=%s", p.String()))
	}
	if u.InsecureTLS == true {
		args = append(args, fmt.Sprintf("--insecure-tls=true"))
	}

	if u.VaultCASecret != "" {
		args = append(args, fmt.Sprintf("--ca-cert-file=%s", vautlCACertFile))

		pt.Spec.Volumes = append(pt.Spec.Volumes, corev1.Volume{
			Name: "vaultCA",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: u.VaultCASecret,
				},
			},
		})

		cont.VolumeMounts = append(cont.VolumeMounts, corev1.VolumeMount{
			Name:      "vaultCA",
			MountPath: filepath.Dir(vautlCACertFile),
		})
	}

	cont.Args = append(cont.Args, args...)

	u.Service.Apply(&cont)

	pt.Spec.Containers = append(pt.Spec.Containers, cont)

	return nil
}

// GetRBAC return rbac roles required by unsealer
func (u *Unsealer) GetRBAC(namespace string, label map[string]string) []rbac.Role {
	return u.Service.GetRBAC(namespace, label)
}
