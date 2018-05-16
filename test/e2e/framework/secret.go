package framework

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "github.com/onsi/gomega"
	"fmt"
)

func (f *Framework) CreateSecret(obj core.Secret) error {
	_, err := f.KubeClient.CoreV1().Secrets(obj.Namespace).Create(&obj)
	return err
}

func (f *Framework) DeleteSecret(meta metav1.ObjectMeta) error {
	return f.KubeClient.CoreV1().Secrets(meta.Namespace).Delete(meta.Name, deleteInForeground())
}

func (f *Framework) EventuallySecret(name, namespace string) GomegaAsyncAssertion {
	return Eventually(func() *core.Secret {
		obj, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		fmt.Println("---secret-----")
		Expect(err).NotTo(HaveOccurred())
		return obj
	},timeOut, pollingInterval)
}