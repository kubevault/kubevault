package framework

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) CreateService(obj core.Service) error {
	_, err := f.KubeClient.CoreV1().Services(obj.Namespace).Create(&obj)
	return err
}

func (f *Framework) DeleteService(meta metav1.ObjectMeta) error {
	return f.KubeClient.CoreV1().Services(meta.Namespace).Delete(meta.Name, deleteInForeground())
}
