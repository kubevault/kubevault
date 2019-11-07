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

package util

import (
	"testing"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsureOwnerRefToObject(t *testing.T) {
	owner := &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vs",
			Namespace: "vs",
			UID:       "1234",
		},
	}

	sPointer := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hi",
			Namespace: "hi",
			UID:       "1234",
		},
	}
	s := core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hi",
			Namespace: "hi",
			UID:       "1234",
		},
	}

	EnsureOwnerRefToObject(sPointer, AsOwner(owner))
	assert.Condition(t, func() (success bool) {
		return IsOwnerRefAlreadyExists(sPointer, AsOwner(owner))
	})

	EnsureOwnerRefToObject(s.GetObjectMeta(), AsOwner(owner))
	assert.Condition(t, func() (success bool) {
		return IsOwnerRefAlreadyExists(s.GetObjectMeta(), AsOwner(owner))
	})
}
