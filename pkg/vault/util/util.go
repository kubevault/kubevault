package util

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultServiceURL returns the DNS record of the vault service in the given namespace.
func VaultServiceURL(name, namespace string, port int) string {
	return fmt.Sprintf("https://%s.%s.svc:%d", name, namespace, port)
}

// ConfigMapNameForVault returns the configMap name for vault
func ConfigMapNameForVault(v *api.VaultServer) string {
	return v.Name + "-vault-config"
}

// LabelsForVault returns the labels for selecting the resources
// belonging to the given vault name.
func LabelsForVault(name string) map[string]string {
	return map[string]string{"app": "vault", "vault_cluster": name}
}

// PodDNSName constructs the dns name on which a pod can be addressed
func PodDNSName(p corev1.Pod) string {
	podIP := strings.Replace(p.Status.PodIP, ".", "-", -1)
	return fmt.Sprintf("%s.%s.pod", podIP, p.Namespace)
}

// AddOwnerRefToObject appends the desired OwnerReference to the object
func AddOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// AsOwner returns an owner reference set as the vault cluster CR
func AsOwner(v *api.VaultServer) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.ResourceKindVaultServer,
		Name:       v.Name,
		UID:        v.UID,
		Controller: &trueVar,
	}
}

func VaultImage(v *api.VaultServer) string {
	return fmt.Sprintf("%s:%s", v.Spec.BaseImage, v.Spec.Version)
}

// image format: baseImageName:tag
func RemoveImageTag(im string) string {
	return strings.Split(im, ":")[0]
}
