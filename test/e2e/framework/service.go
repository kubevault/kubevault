package framework

import (
	core "k8s.io/api/core/v1"
)

func (f *Framework) CreateService(obj core.Service) error {
	_, err := f.KubeClient.CoreV1().Services(obj.Namespace).Create(&obj)
	return err
}

func (f *Framework) DeleteService(name, namespace string) error {
	return f.KubeClient.CoreV1().Services(namespace).Delete(name, deleteInForeground())
}
