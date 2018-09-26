package util

import (
	"fmt"
	"strings"

	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultServiceURL returns the DNS record of the vault service in the given namespace.
func VaultServiceURL(name, namespace string, port int) string {
	return fmt.Sprintf("https://%s.%s.svc:%d", name, namespace, port)
}

// PodDNSName constructs the dns name on which a pod can be addressed
func PodDNSName(p core.Pod) string {
	podIP := strings.Replace(p.Status.PodIP, ".", "-", -1)
	return fmt.Sprintf("%s.%s.pod", podIP, p.Namespace)
}

// EnsureOwnerRefToObject appends the desired OwnerReference to the object
func EnsureOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	if !IsOwnerRefAlreadyExists(o, r) {
		o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
	}
}

// IsOwnerRefAlreadyExists checks whether owner ref already exists
func IsOwnerRefAlreadyExists(o metav1.Object, r metav1.OwnerReference) bool {
	refs := o.GetOwnerReferences()
	for _, u := range refs {
		if u.Name == r.Name &&
			u.UID == r.UID &&
			u.Kind == r.Kind &&
			u.APIVersion == u.APIVersion {
			return true
		}
	}
	return false
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
