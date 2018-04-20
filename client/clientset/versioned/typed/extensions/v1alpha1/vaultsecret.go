/*
Copyright 2018 The Vault Operator Authors.

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

package v1alpha1

import (
	v1alpha1 "github.com/soter/vault-operator/apis/extensions/v1alpha1"
	scheme "github.com/soter/vault-operator/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	rest "k8s.io/client-go/rest"
)

// VaultSecretsGetter has a method to return a VaultSecretInterface.
// A group's client should implement this interface.
type VaultSecretsGetter interface {
	VaultSecrets(namespace string) VaultSecretInterface
}

// VaultSecretInterface has methods to work with VaultSecret resources.
type VaultSecretInterface interface {
	Create(*v1alpha1.VaultSecret) (*v1alpha1.VaultSecret, error)
	Update(*v1alpha1.VaultSecret) (*v1alpha1.VaultSecret, error)
	UpdateStatus(*v1alpha1.VaultSecret) (*v1alpha1.VaultSecret, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.VaultSecret, error)
	List(opts v1.ListOptions) (*v1alpha1.VaultSecretList, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VaultSecret, err error)
	VaultSecretExpansion
}

// vaultSecrets implements VaultSecretInterface
type vaultSecrets struct {
	client rest.Interface
	ns     string
}

// newVaultSecrets returns a VaultSecrets
func newVaultSecrets(c *ExtensionsV1alpha1Client, namespace string) *vaultSecrets {
	return &vaultSecrets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the vaultSecret, and returns the corresponding vaultSecret object, and an error if there is any.
func (c *vaultSecrets) Get(name string, options v1.GetOptions) (result *v1alpha1.VaultSecret, err error) {
	result = &v1alpha1.VaultSecret{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vaultsecrets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VaultSecrets that match those selectors.
func (c *vaultSecrets) List(opts v1.ListOptions) (result *v1alpha1.VaultSecretList, err error) {
	result = &v1alpha1.VaultSecretList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vaultsecrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Create takes the representation of a vaultSecret and creates it.  Returns the server's representation of the vaultSecret, and an error, if there is any.
func (c *vaultSecrets) Create(vaultSecret *v1alpha1.VaultSecret) (result *v1alpha1.VaultSecret, err error) {
	result = &v1alpha1.VaultSecret{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("vaultsecrets").
		Body(vaultSecret).
		Do().
		Into(result)
	return
}

// Update takes the representation of a vaultSecret and updates it. Returns the server's representation of the vaultSecret, and an error, if there is any.
func (c *vaultSecrets) Update(vaultSecret *v1alpha1.VaultSecret) (result *v1alpha1.VaultSecret, err error) {
	result = &v1alpha1.VaultSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("vaultsecrets").
		Name(vaultSecret.Name).
		Body(vaultSecret).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *vaultSecrets) UpdateStatus(vaultSecret *v1alpha1.VaultSecret) (result *v1alpha1.VaultSecret, err error) {
	result = &v1alpha1.VaultSecret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("vaultsecrets").
		Name(vaultSecret.Name).
		SubResource("status").
		Body(vaultSecret).
		Do().
		Into(result)
	return
}

// Delete takes name of the vaultSecret and deletes it. Returns an error if one occurs.
func (c *vaultSecrets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vaultsecrets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *vaultSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vaultsecrets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched vaultSecret.
func (c *vaultSecrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VaultSecret, err error) {
	result = &v1alpha1.VaultSecret{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("vaultsecrets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
