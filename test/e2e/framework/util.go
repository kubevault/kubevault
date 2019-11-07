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
	vaultapi "github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteInBackground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func deleteInForeground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func EnsureKubernetesAuth(vc *vaultapi.Client) error {
	authList, err := vc.Sys().ListAuth()
	if err != nil {
		return err
	}
	for _, v := range authList {
		if v.Type == "kubernetes" {
			// kubernetes auth already enabled
			return nil
		}
	}

	err = vc.Sys().EnableAuthWithOptions("kubernetes", &vaultapi.EnableAuthOptions{
		Type: "kubernetes",
	})
	return err
}
