package framework

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func (f *Framework) CreateAppBinding(a *appcat.AppBinding) error {
	_, err := f.AppcatClient.AppBindings(a.Namespace).Create(a)
	return err
}

func (f *Framework) GetAppBinding(name, namespace string) (*appcat.AppBinding, error) {
	return f.AppcatClient.AppBindings(namespace).Get(name, metav1.GetOptions{})
}

func (f *Framework) DeleteAppBinding(name, namespace string) error {
	return f.AppcatClient.AppBindings(namespace).Delete(name, deleteInForeground())
}
