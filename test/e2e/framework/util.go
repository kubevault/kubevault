package framework

import (
	"fmt"

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

func GetVaultClient(addr, port string) (*vaultapi.Client, error) {
	cfg := vaultapi.DefaultConfig()
	podURL := fmt.Sprintf("https://%s:%s", addr, port)
	cfg.Address = podURL
	cfg.ConfigureTLS(nil)
	return vaultapi.NewClient(cfg)
}
