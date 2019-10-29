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
	"fmt"

	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
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
	err := f.KubeClient.CoreV1().Secrets(namespace).Delete(name, deleteInForeground())
	if kerr.IsNotFound(err) {
		return nil
	}
	return err
}

func (f *Framework) EventuallySecret(name, namespace string) GomegaAsyncAssertion {
	return Eventually(func() *core.Secret {
		obj, err := f.KubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		fmt.Println("---secret-----")
		Expect(err).NotTo(HaveOccurred())
		return obj
	}, timeOut, pollingInterval)
}
