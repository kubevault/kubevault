/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

func (f *Framework) CreateAppBinding(a *appcat.AppBinding) error {
	_, err := f.AppcatClient.AppBindings(a.Namespace).Create(context.TODO(), a, metav1.CreateOptions{})
	return err
}

func (f *Framework) GetAppBinding(name, namespace string) (*appcat.AppBinding, error) {
	return f.AppcatClient.AppBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (f *Framework) DeleteAppBinding(name, namespace string) error {
	return f.AppcatClient.AppBindings(namespace).Delete(context.TODO(), name, meta_util.DeleteInForeground())
}

func (f *Framework) CreateLocalRef2AppRef(namespace string, reference *v1.LocalObjectReference) *appcat.AppReference {
	return &appcat.AppReference{
		Namespace: namespace,
		Name:      reference.Name,
	}
}
