package util

import (
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#token-controller
func GetJwtTokenSecretFromServiceAccount(kc kubernetes.Interface, name, namespace string) (*core.Secret, error) {
	sa, err := kc.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(sa.Secrets) == 0 {
		return nil, errors.New("token secret still haven't created yet")
	} else {
		// get secret
		for _, s := range sa.Secrets {
			sr, err := kc.CoreV1().Secrets(namespace).Get(s.Name, metav1.GetOptions{})
			if err == nil {
				return sr, nil
			} else if !kerr.IsNotFound(err) {
				return nil, err
			}
		}
		return nil, errors.New("token secret is not available")
	}
}

func TryGetJwtTokenSecretNameFromServiceAccount(kc kubernetes.Interface, name string, namespace string, interval time.Duration, timeout time.Duration) (*core.Secret, error) {
	var (
		err    error
		secret *core.Secret
	)
	err2 := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		secret, err = GetJwtTokenSecretFromServiceAccount(kc, name, namespace)
		if err == nil {
			return true, nil
		} else {
			glog.Errorf("trying to get jwt token secret name from service account %s/%s: %s", namespace, name, err)
		}
		return false, nil
	})
	if err2 != nil {
		return nil, errors.Wrap(err, err2.Error())
	}
	return secret, nil
}
