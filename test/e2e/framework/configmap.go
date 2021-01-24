/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"

	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (f *Framework) CreateConfigMap(obj core.ConfigMap) error {
	_, err := f.KubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(context.TODO(), &obj, metav1.CreateOptions{})
	return err
}

func (f *Framework) DeleteConfigMap(meta metav1.ObjectMeta) error {
	return f.KubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(context.TODO(), meta.Name, meta_util.DeleteInForeground())
}

func (f *Framework) EventuallyConfigMap(name, namespace string) GomegaAsyncAssertion {
	return Eventually(func() *core.ConfigMap {
		obj, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
