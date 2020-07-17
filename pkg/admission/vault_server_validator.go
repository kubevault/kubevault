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

package admission

import (
	"context"
	"fmt"
	"strings"
	"sync"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	cs "kubevault.dev/operator/client/clientset/versioned"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
)

const (
	validatorGroup   = "validators.kubevault.com"
	validatorVersion = "v1alpha1"
)

type VaultServerValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &VaultServerValidator{}

func (v *VaultServerValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    validatorGroup,
			Version:  validatorVersion,
			Resource: "vaultservervalidators",
		},
		"vaultservervalidator"
}

func (v *VaultServerValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.initialized = true

	var err error
	if v.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	if v.extClient, err = cs.NewForConfig(config); err != nil {
		return err
	}
	return err
}

func (v *VaultServerValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindVaultServer {
		status.Allowed = true
		return status
	}

	v.lock.RLock()
	defer v.lock.RUnlock()
	if !v.initialized {
		return hookapi.StatusUninitialized()
	}

	if req.Operation == admission.Create || req.Operation == admission.Update {
		obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		if req.Operation == admission.Update {
			// validate changes made by user
			oldObject, err := meta_util.UnmarshalFromJSON(req.OldObject.Raw, api.SchemeGroupVersion)
			if err != nil {
				return hookapi.StatusBadRequest(err)
			}

			vs := obj.(*api.VaultServer).DeepCopy()
			oldVs := oldObject.(*api.VaultServer).DeepCopy()

			if err := validateUpdate(vs, oldVs); err != nil {
				return hookapi.StatusBadRequest(err)
			}
		}
		// validate vaultserver specs
		if err = ValidateVaultServer(v.client, v.extClient, obj.(*api.VaultServer)); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// ValidateVaultServer checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateVaultServer(client kubernetes.Interface, extClient cs.Interface, vs *api.VaultServer) error {
	if vs.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	if _, err := extClient.CatalogV1alpha1().VaultServerVersions().Get(context.TODO(), string(vs.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if vs.Spec.Replicas != nil && *vs.Spec.Replicas < 1 {
		return errors.Errorf(`spec.nodes "%v" invalid. Value must be greater than zero`, vs.Spec.Replicas)
	}

	numOfBackend := 0
	if vs.Spec.Backend.Inmem != nil {
		numOfBackend++
	}
	if vs.Spec.Backend.File != nil {
		numOfBackend++
	}
	if vs.Spec.Backend.Etcd != nil {
		numOfBackend++
		etcd := vs.Spec.Backend.Etcd
		if etcd.CredentialSecretName != "" {
			err := validateSecret(client, etcd.CredentialSecretName, vs.Namespace, []string{
				"username",
				"password",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.etcd.credentialSecretName")
			}
		}
		if etcd.TLSSecretName != "" {
			err := validateSecret(client, etcd.TLSSecretName, vs.Namespace, []string{
				"ca.crt",
				"client.crt",
				"client.key",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.etcd.tlsSecretName")
			}
		}
	}
	if vs.Spec.Backend.MySQL != nil {
		numOfBackend++
		my := vs.Spec.Backend.MySQL
		if my.UserCredentialSecret != "" {
			err := validateSecret(client, my.UserCredentialSecret, vs.Namespace, []string{
				"username",
				"password",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.mySQL.userCredentialSecret")
			}
		}
		if my.TLSCASecret != "" {
			err := validateSecret(client, my.TLSCASecret, vs.Namespace, []string{
				"tls_ca_file",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.mySQL.tlsCASecret")
			}
		}
	}
	if vs.Spec.Backend.PostgreSQL != nil {
		numOfBackend++
		pg := vs.Spec.Backend.PostgreSQL
		if pg.ConnectionURLSecret != "" {
			err := validateSecret(client, pg.ConnectionURLSecret, vs.Namespace, []string{
				"connection_url",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.postgreSQL.connectionURLSecret")
			}
		}
	}
	if vs.Spec.Backend.Gcs != nil {
		numOfBackend++
		gcs := vs.Spec.Backend.Gcs
		if gcs.CredentialSecret != "" {
			err := validateSecret(client, gcs.CredentialSecret, vs.Namespace, []string{
				"sa.json",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.gcs.credentialSecret")
			}
		}
	}
	if vs.Spec.Backend.S3 != nil {
		numOfBackend++
		s3 := vs.Spec.Backend.S3
		if s3.CredentialSecret != "" {
			err := validateSecret(client, s3.CredentialSecret, vs.Namespace, []string{
				"access_key",
				"secret_key",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.s3.credentialSecret")
			}
		}
		if s3.SessionTokenSecret != "" {
			err := validateSecret(client, s3.SessionTokenSecret, vs.Namespace, []string{
				"session_token",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.s3.sessionTokenSecret")
			}
		}
	}
	if vs.Spec.Backend.Azure != nil {
		numOfBackend++
		azure := vs.Spec.Backend.Azure
		if azure.AccountKeySecret != "" {
			err := validateSecret(client, azure.AccountKeySecret, vs.Namespace, []string{
				"account_key",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.azure.accountKeySecret")
			}
		}
	}
	if vs.Spec.Backend.DynamoDB != nil {
		numOfBackend++
		dyndb := vs.Spec.Backend.DynamoDB
		if dyndb.CredentialSecret != "" {
			err := validateSecret(client, dyndb.CredentialSecret, vs.Namespace, []string{
				"access_key",
				"secret_key",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.dynamoDB.credentialSecret")
			}
		}
		if dyndb.SessionTokenSecret != "" {
			err := validateSecret(client, dyndb.SessionTokenSecret, vs.Namespace, []string{
				"session_token",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.dynamoDB.sessionTokenSecret")
			}
		}
	}
	if vs.Spec.Backend.Swift != nil {
		numOfBackend++
		swft := vs.Spec.Backend.Swift
		if swft.CredentialSecret != "" {
			err := validateSecret(client, swft.CredentialSecret, vs.Namespace, []string{
				"username",
				"password",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.swift.credentialSecret")
			}
		}
		if swft.AuthTokenSecret != "" {
			err := validateSecret(client, swft.AuthTokenSecret, vs.Namespace, []string{
				"auth_token",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.swift.authTokenSecret")
			}
		}
	}

	if vs.Spec.Backend.Consul != nil {
		numOfBackend++
		consul := vs.Spec.Backend.Consul
		if consul.ACLTokenSecretName != "" {
			err := validateSecret(client, consul.ACLTokenSecretName, vs.Namespace, []string{
				"aclToken",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.consul.aclTokenSecretName")
			}
		}
		if consul.TLSSecretName != "" {
			err := validateSecret(client, consul.TLSSecretName, vs.Namespace, []string{
				"ca.crt",
				"client.crt",
				"client.key",
			})
			if err != nil {
				return errors.Wrap(err, "for spec.backend.consul.tlsSecretName")
			}
		}
	}

	if vs.Spec.Backend.Raft != nil {
		numOfBackend++
	}

	if numOfBackend != 1 {
		if numOfBackend == 0 {
			return errors.New("spec.backend is not specified")
		} else if numOfBackend > 1 {
			return errors.New("more than one spec.backend is specified")
		}
	}

	if vs.Spec.Unsealer != nil {
		unslr := vs.Spec.Unsealer
		if unslr.SecretShares <= 0 {
			return errors.New("spec.unsealer.secretShares must be greater than zero")
		}
		if unslr.SecretThreshold <= 0 {
			return errors.New("spec.unsealer.secretThreshold must be greater than zero")
		}
		if unslr.SecretThreshold > unslr.SecretShares {
			return errors.New("spec.unsealer.secretShares must be greater than spec.unsealer.secretThreshold")
		}

		mode := unslr.Mode
		numOfModes := 0
		if mode.KubernetesSecret != nil {
			numOfModes++
		}
		if mode.GoogleKmsGcs != nil {
			numOfModes++
			if mode.GoogleKmsGcs.CredentialSecret != "" {
				err := validateSecret(client, mode.GoogleKmsGcs.CredentialSecret, vs.Namespace, []string{
					"sa.json",
				})
				if err != nil {
					return errors.Wrap(err, "for spec.unsealer.mode.googleKmsGcs.credentialSecret")
				}
			}
		}
		if mode.AwsKmsSsm != nil {
			numOfModes++
			if mode.AwsKmsSsm.CredentialSecret != "" {
				err := validateSecret(client, mode.AwsKmsSsm.CredentialSecret, vs.Namespace, []string{
					"access_key",
					"secret_key",
				})
				if err != nil {
					return errors.Wrap(err, "for spec.unsealer.mode.awsKmsSsm.credentialSecret")
				}
			}
		}
		if mode.AzureKeyVault != nil {
			numOfModes++
			akv := mode.AzureKeyVault
			if akv.AADClientSecret != "" {
				err := validateSecret(client, akv.AADClientSecret, vs.Namespace, []string{
					"client-id",
					"client-secret",
				})
				if err != nil {
					return errors.Wrap(err, "for spec.unsealer.mode.awsKmsSsm.aadClientSecret")
				}
			}
			if akv.ClientCertSecret != "" {
				err := validateSecret(client, akv.ClientCertSecret, vs.Namespace, []string{
					"client-cert",
					"client-cert-password",
				})
				if err != nil {
					return errors.Wrap(err, "for spec.unsealer.mode.awsKmsSsm.clientCertSecret")
				}
			}

		}

		if numOfModes != 1 {
			if numOfModes == 0 {
				return errors.New("spec.unsealer.mode is not specified")
			} else if numOfModes > 1 {
				return errors.New("more than one spec.unsealer.mode is specified")
			}
		}

	}
	return nil
}

func validateUpdate(obj, oldObj runtime.Object) error {
	preconditions := getPreconditionFunc()
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError())
		}
		return err
	}
	return nil
}

func getPreconditionFunc() []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
	}

	for _, field := range preconditionSpecFields {
		preconditions = append(preconditions,
			meta_util.RequireChainKeyUnchanged(field),
		)
	}
	return preconditions
}

var preconditionSpecFields = []string{
	"spec.unsealer",
	"spec.backend",
	"spec.podTemplate.spec.nodeSelector",
}

func preconditionFailedError() error {
	str := preconditionSpecFields
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}

// validateSecret will check:
//	- whether secret exists
//	- whether value for requiredKeys exists
func validateSecret(kc kubernetes.Interface, name string, ns string, requiredKeys []string) error {
	sr, err := kc.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, k := range requiredKeys {
		val, ok := sr.Data[k]
		if !ok || len(val) == 0 {
			return errors.Errorf("secret data doesn't contain any value for key '%s'", k)
		}
	}
	return nil
}
