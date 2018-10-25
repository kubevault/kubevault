package unsealer

import (
	"fmt"
	"time"

	core_util "github.com/appscode/kutil/core/v1"
	"github.com/golang/glog"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	sa_util "github.com/kubevault/operator/pkg/util"
	"github.com/kubevault/operator/pkg/vault/unsealer/aws"
	"github.com/kubevault/operator/pkg/vault/unsealer/azure"
	"github.com/kubevault/operator/pkg/vault/unsealer/google"
	k8s "github.com/kubevault/operator/pkg/vault/unsealer/kubernetes"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	K8sTokenReviewerJwtEnv = "K8S_TOKEN_REVIEWER_JWT"
)

type Unsealer interface {
	Apply(pt *core.PodTemplateSpec) error
	GetRBAC(prefix, namespace string) []rbac.Role
}

type unsealerSrv struct {
	restConfig *rest.Config
	kc         kubernetes.Interface
	vs         *api.VaultServer
	unsealer   Unsealer
	image      string
}

func newUnsealer(s *api.UnsealerSpec) (Unsealer, error) {
	if s.Mode.KubernetesSecret != nil {
		return k8s.NewOptions(*s.Mode.KubernetesSecret)
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

func NewUnsealerService(restConfig *rest.Config, vs *api.VaultServer, image string) (Unsealer, error) {
	if vs == nil {
		return nil, errors.New("VaultServer is nil")
	}
	if vs.Spec.Unsealer == nil {
		glog.Infoln(".spec.unsealer is nil")
		return nil, nil
	}

	unslr, err := newUnsealer(vs.Spec.Unsealer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create unsealer service")
	}

	kc, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes client")
	}
	return &unsealerSrv{
		restConfig: restConfig,
		vs:         vs,
		kc:         kc,
		image:      image,
		unsealer:   unslr,
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

	unslr := u.vs.Spec.Unsealer
	cont := core.Container{
		Name:  util.VaultUnsealerContainerName,
		Image: u.image,
	}
	args = append(args,
		"run",
		"--v=3",
	)

	if unslr.SecretShares != 0 {
		args = append(args, fmt.Sprintf("--secret-shares=%d", unslr.SecretShares))
	}
	if unslr.SecretThreshold != 0 {
		args = append(args, fmt.Sprintf("--secret-threshold=%d", unslr.SecretThreshold))
	}

	if unslr.RetryPeriodSeconds != 0 {
		p := time.Second * unslr.RetryPeriodSeconds
		args = append(args, fmt.Sprintf("--retry-period=%s", p.String()))
	}
	if unslr.InsecureSkipTLSVerify == true {
		args = append(args, fmt.Sprintf("--insecure-skip-tls-verify=true"))
	}
	if unslr.OverwriteExisting == true {
		args = append(args, fmt.Sprintf("--overwrite-existing=true"))
	}

	if unslr.InsecureSkipTLSVerify == false && len(unslr.CABundle) != 0 {
		args = append(args, fmt.Sprintf("--ca-cert=%s", unslr.CABundle))
	}

	// Add kubernetes auth flags
	args = append(args, fmt.Sprintf("--auth.k8s-host=%s", u.restConfig.Host))

	err := rest.LoadTLSFiles(u.restConfig)
	if err != nil {
		return errors.Wrap(err, "fialed to TLS files from rest config for kubernetes auth")
	}
	args = append(args, fmt.Sprintf("--auth.k8s-ca-cert=%s", u.restConfig.CAData))

	// Add env for kubernetes auth
	secretName, err := sa_util.TryGetJwtTokenSecretNameFromServiceAccount(u.kc, u.vs.ServiceAccountForTokenReviewer(), u.vs.Namespace, 2*time.Second, 30*time.Second)
	if err != nil {
		return errors.Wrapf(err, "failed to get jwt token secret name of service account(%s/%s)", u.vs.Namespace, u.vs.ServiceAccountForTokenReviewer())
	}
	cont.Env = core_util.UpsertEnvVars(cont.Env, core.EnvVar{
		Name: K8sTokenReviewerJwtEnv,
		ValueFrom: &core.EnvVarSource{
			SecretKeyRef: &core.SecretKeySelector{
				LocalObjectReference: core.LocalObjectReference{
					Name: secretName,
				},
				Key: core.ServiceAccountTokenKey,
			},
		},
	})

	// Add flags for policy
	args = append(args, fmt.Sprintf("--policy.name=%s", u.vs.PolicyNameForPolicyController()))
	args = append(args, fmt.Sprintf("--policy.service-account-name=%s", u.vs.ServiceAccountForPolicyController()))
	args = append(args, fmt.Sprintf("--policy.service-account-namespace=%s", u.vs.Namespace))

	cont.Args = append(cont.Args, args...)
	pt.Spec.Containers = core_util.UpsertContainer(pt.Spec.Containers, cont)
	err = u.unsealer.Apply(pt)
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
