package util

import (
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func GetJwtTokenSecretNameFromServiceAccount(kc kubernetes.Interface, name, namespace string) (string, error) {
	sa, err := kc.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if len(sa.Secrets) == 0 {
		return "", errors.New("token secret still haven't created yet")
	} else {
		return sa.Secrets[0].Name, nil
	}
}

func TryGetJwtTokenSecretNameFromServiceAccount(kc kubernetes.Interface, name string, namespace string, interval time.Duration, timeout time.Duration) (string, error) {
	var (
		err        error
		secretName string
	)
	err2 := wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		secretName, err = GetJwtTokenSecretNameFromServiceAccount(kc, name, namespace)
		if err == nil {
			return true, nil
		} else {
			glog.Errorf("trying to get jwt token secret name from service account %s/%s: %s", namespace, name, err)
		}
		return false, nil
	})
	if err2 != nil {
		return "", errors.Wrap(err, err2.Error())
	}
	return secretName, err2
}
