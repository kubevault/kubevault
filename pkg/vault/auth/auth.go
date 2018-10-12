package auth

import (
	awsauth "github.com/kubevault/operator/pkg/vault/auth/aws"
	certauth "github.com/kubevault/operator/pkg/vault/auth/cert"
	k8sauth "github.com/kubevault/operator/pkg/vault/auth/kubernetes"
	tokenauth "github.com/kubevault/operator/pkg/vault/auth/token"
	basicauth "github.com/kubevault/operator/pkg/vault/auth/userpass"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type AuthInterface interface {
	// successful login will return token
	// unsuccessful login will return err
	Login() (string, error)
}

func NewAuth(kc kubernetes.Interface, vApp *appcat.AppBinding) (AuthInterface, error) {
	if vApp == nil {
		return nil, errors.New("vault AppBinding is not provided")
	}
	if vApp.Spec.Secret == nil {
		return nil, errors.New("secret is not provided")
	}

	secret, err := kc.CoreV1().Secrets(vApp.Namespace).Get(vApp.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if secret.Type == "kubernetes.io/basic-auth" {
		return basicauth.New(vApp, secret)
	} else if secret.Type == "kubevault.com/service-account" {
		return k8sauth.New(vApp, secret)
	} else if secret.Type == "kubevault.com/token" {
		return tokenauth.New(secret)
	} else if secret.Type == "kubevault.com/aws" {
		return awsauth.New(vApp, secret)
	} else if secret.Type == "kubernetes.io/tls" {
		return certauth.New(vApp, secret)
	} else {
		return nil, errors.New("Invalid/Unsupported secret type")
	}
}
