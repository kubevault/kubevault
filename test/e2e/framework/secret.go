package framework

import (
	"fmt"

	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) CreateSecret(obj core.Secret) error {
	_, err := f.KubeClient.CoreV1().Secrets(obj.Namespace).Create(&obj)
	return err
}

func (f *Framework) CreateSecretWithData(name, namespace string, data map[string][]byte) error {
	sr := core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	err := f.CreateSecret(sr)
	return err
}

func (f *Framework) DeleteSecret(name, namespace string) error {
	return f.KubeClient.CoreV1().Secrets(namespace).Delete(name, deleteInForeground())
}

func (f *Framework) EventuallySecret(name, namespace string) GomegaAsyncAssertion {
	return Eventually(func() *core.Secret {
		obj, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		fmt.Println("---secret-----")
		Expect(err).NotTo(HaveOccurred())
		return obj
	}, timeOut, pollingInterval)
}
