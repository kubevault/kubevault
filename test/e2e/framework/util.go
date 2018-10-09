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
