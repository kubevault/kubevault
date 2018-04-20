package main

import (
	"io/ioutil"
	"os"

	"github.com/appscode/go/log"
	gort "github.com/appscode/go/runtime"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/openapi"
	"github.com/go-openapi/spec"
	"github.com/golang/glog"
	extinstall "github.com/soter/vault-operator/apis/extensions/install"
	repov1alpha1 "github.com/soter/vault-operator/apis/extensions/v1alpha1"
	vaultinstall "github.com/soter/vault-operator/apis/vault/install"
	stashv1alpha1 "github.com/soter/vault-operator/apis/vault/v1alpha1"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kube-openapi/pkg/common"
)

func generateCRDDefinitions() {
	filename := gort.GOPath() + "/src/github.com/soter/vault-operator/apis/vault/v1alpha1/crds.yaml"

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	crds := []*crd_api.CustomResourceDefinition{
		stashv1alpha1.VaultServer{}.CustomResourceDefinition(),
	}
	for _, crd := range crds {
		crdutils.MarshallCrd(f, crd, "yaml")
	}
}
func generateSwaggerJson() {
	var (
		groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
		registry             = registered.NewOrDie("")
		Scheme               = runtime.NewScheme()
		Codecs               = serializer.NewCodecFactory(Scheme)
	)

	vaultinstall.Install(groupFactoryRegistry, registry, Scheme)
	extinstall.Install(groupFactoryRegistry, registry, Scheme)

	apispec, err := openapi.RenderOpenAPISpec(openapi.Config{
		Registry: registry,
		Scheme:   Scheme,
		Codecs:   Codecs,
		Info: spec.InfoProps{
			Title:   "Vault",
			Version: "v0",
			Contact: &spec.ContactInfo{
				Name:  "AppsCode Inc.",
				URL:   "https://appscode.com",
				Email: "hello@appscode.com",
			},
			License: &spec.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		OpenAPIDefinitions: []common.GetOpenAPIDefinitions{
			stashv1alpha1.GetOpenAPIDefinitions,
			repov1alpha1.GetOpenAPIDefinitions,
		},
		Resources: []schema.GroupVersionResource{
			stashv1alpha1.SchemeGroupVersion.WithResource(stashv1alpha1.ResourcePluralVaultServer),
		},
		RDResources: []schema.GroupVersionResource{
			repov1alpha1.SchemeGroupVersion.WithResource(repov1alpha1.ResourcePluralVaultSecret),
		},
	})
	if err != nil {
		glog.Fatal(err)
	}

	filename := gort.GOPath() + "/src/github.com/soter/vault-operator/openapi-spec/v2/swagger.json"
	err = ioutil.WriteFile(filename, []byte(apispec), 0644)
	if err != nil {
		glog.Fatal(err)
	}
}

func main() {
	generateCRDDefinitions()
	generateSwaggerJson()
}
