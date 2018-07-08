package server

import (
	"fmt"
	"io"
	"net"

	"github.com/kubevault/operator/apis/extensions/v1alpha1"
	"github.com/kubevault/operator/pkg/controller"
	"github.com/kubevault/operator/pkg/server"
	"github.com/spf13/pflag"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/core.kubevault.com"

type StashOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	ExtraOptions       *ExtraOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewStashOptions(out, errOut io.Writer) *StashOptions {
	o := &StashOptions{
		// TODO we will nil out the etcd storage options.  This requires a later level of k8s.io/apiserver
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, server.Codecs.LegacyCodec(admissionv1beta1.SchemeGroupVersion)),
		ExtraOptions:       NewExtraOptions(),
		StdOut:             out,
		StdErr:             errOut,
	}
	o.RecommendedOptions.Etcd = nil

	return o
}

func (o StashOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
	o.ExtraOptions.AddFlags(fs)
}

func (o StashOptions) Validate(args []string) error {
	return nil
}

func (o *StashOptions) Complete() error {
	return nil
}

func (o StashOptions) Config() (*server.StashConfig, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(server.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig, server.Scheme); err != nil {
		return nil, err
	}
	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(v1alpha1.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(server.Scheme))
	serverConfig.OpenAPIConfig.Info.Title = "vault-operator"
	serverConfig.OpenAPIConfig.Info.Version = v1alpha1.SchemeGroupVersion.Version
	serverConfig.OpenAPIConfig.IgnorePrefixes = []string{
		"/swaggerapi",
		"/apis/admission.kubevault.com/v1alpha1/vaultservers",
		"/apis/admission.kubevault.com/v1alpha1/deployments",
		"/apis/admission.kubevault.com/v1alpha1/daemonsets",
		"/apis/admission.kubevault.com/v1alpha1/statefulsets",
		"/apis/admission.kubevault.com/v1alpha1/replicationcontrollers",
		"/apis/admission.kubevault.com/v1alpha1/replicasets",
		"/apis/admission.kubevault.com/v1alpha1/jobs",
		"/apis/admission.kubevault.com/v1alpha1/cronjobs",
	}

	extraConfig := controller.NewConfig(serverConfig.ClientConfig)
	if err := o.ExtraOptions.ApplyTo(extraConfig); err != nil {
		return nil, err
	}

	config := &server.StashConfig{
		GenericConfig: serverConfig,
		ExtraConfig:   extraConfig,
	}
	return config, nil
}

func (o StashOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	s, err := config.Complete().New()
	if err != nil {
		return err
	}

	return s.Run(stopCh)
}
