package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/appscode/go/log"
	gort "github.com/appscode/go/runtime"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/openapi"
	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	"github.com/kubevault/operator/apis"
	cataloginstall "github.com/kubevault/operator/apis/catalog/install"
	catalogv1alpha1 "github.com/kubevault/operator/apis/catalog/v1alpha1"
	secretinstall "github.com/kubevault/operator/apis/engine/install"
	secretv1alpha1 "github.com/kubevault/operator/apis/engine/v1alpha1"
	vaultinstall "github.com/kubevault/operator/apis/kubevault/install"
	vaultv1alpha1 "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	policyinstall "github.com/kubevault/operator/apis/policy/install"
	policyv1alpha1 "github.com/kubevault/operator/apis/policy/v1alpha1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
)

func generateCRDDefinitions() {
	apis.EnableStatusSubresource = true

	filename := gort.GOPath() + "/src/github.com/kubevault/operator/apis/kubevault/v1alpha1/crds.yaml"
	os.Remove(filename)

	err := os.MkdirAll(filepath.Join(gort.GOPath(), "/src/github.com/kubevault/operator/api/crds"), 0755)
	if err != nil {
		log.Fatal(err)
	}

	crds := []*crd_api.CustomResourceDefinition{
		vaultv1alpha1.VaultServer{}.CustomResourceDefinition(),
		catalogv1alpha1.VaultServerVersion{}.CustomResourceDefinition(),
		policyv1alpha1.VaultPolicy{}.CustomResourceDefinition(),
		policyv1alpha1.VaultPolicyBinding{}.CustomResourceDefinition(),
		secretv1alpha1.AWSRole{}.CustomResourceDefinition(),
		secretv1alpha1.AWSAccessKeyRequest{}.CustomResourceDefinition(),
	}
	for _, crd := range crds {
		filename := filepath.Join(gort.GOPath(), "/src/github.com/kubevault/operator/api/crds", crd.Spec.Names.Singular+".yaml")
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
		crdutils.MarshallCrd(f, crd, "yaml")
		f.Close()
	}
}

func generateSwaggerJson() {
	var (
		Scheme = runtime.NewScheme()
		Codecs = serializer.NewCodecFactory(Scheme)
	)

	vaultinstall.Install(Scheme)
	cataloginstall.Install(Scheme)
	policyinstall.Install(Scheme)
	secretinstall.Install(Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Scheme: Scheme,
		Codecs: Codecs,
		Info: spec.InfoProps{
			Title:   "KubeVault",
			Version: "v0.1.0",
			Contact: &spec.ContactInfo{
				Name:  "AppsCode Inc.",
				URL:   "https://appscode.com",
				Email: "kubevault@appscode.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			vaultv1alpha1.GetOpenAPIDefinitions,
			catalogv1alpha1.GetOpenAPIDefinitions,
			policyv1alpha1.GetOpenAPIDefinitions,
			secretv1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []openapi.TypeInfo{
			{vaultv1alpha1.SchemeGroupVersion, vaultv1alpha1.ResourceVaultServers, vaultv1alpha1.ResourceKindVaultServer, true},
			{catalogv1alpha1.SchemeGroupVersion, catalogv1alpha1.ResourceVaultServerVersions, catalogv1alpha1.ResourceKindVaultServerVersion, false},
			{policyv1alpha1.SchemeGroupVersion, policyv1alpha1.ResourceVaultPolicies, policyv1alpha1.ResourceKindVaultPolicy, true},
			{policyv1alpha1.SchemeGroupVersion, policyv1alpha1.ResourceVaultPolicyBindings, policyv1alpha1.ResourceKindVaultPolicyBinding, true},
			{secretv1alpha1.SchemeGroupVersion, secretv1alpha1.ResourceAWSRoles, secretv1alpha1.ResourceKindAWSRole, true},
			{secretv1alpha1.SchemeGroupVersion, secretv1alpha1.ResourceAWSAccessKeyRequests, secretv1alpha1.ResourceKindAWSAccessKeyRequest, true},
		},
	})
	if err != nil {
		glog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/github.com/kubevault/operator/api/openapi-spec/swagger.json"
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		glog.Fatal(err)
	}
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		glog.Fatal(err)
	}
}

func main() {
	generateCRDDefinitions()
	generateSwaggerJson()
}
