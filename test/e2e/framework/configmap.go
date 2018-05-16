package framework

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "github.com/onsi/gomega"
)

func (f *Framework) CreateConfigMap(obj core.ConfigMap) error {
	_, err := f.KubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(&obj)
	return err
}

func (f *Framework) DeleteConfigMap(meta metav1.ObjectMeta) error {
	return f.KubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(meta.Name, deleteInForeground())
}

func (f *Framework) EventuallyConfigMap(name, namespace string) GomegaAsyncAssertion {
	return Eventually(func() *core.ConfigMap {
		obj, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}