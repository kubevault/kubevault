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

// SecretsGetter has a method to return a SecretInterface.
// A group's client should implement this interface.
type SecretsGetter interface {
	Secrets(namespace string) SecretInterface
}

// SecretInterface has methods to work with Secret resources.
type SecretInterface interface {
	Create(*v1alpha1.Secret) (*v1alpha1.Secret, error)
	Update(*v1alpha1.Secret) (*v1alpha1.Secret, error)
	UpdateStatus(*v1alpha1.Secret) (*v1alpha1.Secret, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Secret, error)
	List(opts v1.ListOptions) (*v1alpha1.SecretList, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Secret, err error)
	SecretExpansion
}

// secrets implements SecretInterface
type secrets struct {
	client rest.Interface
	ns     string
}

// newSecrets returns a Secrets
func newSecrets(c *ExtensionsV1alpha1Client, namespace string) *secrets {
	return &secrets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the secret, and returns the corresponding secret object, and an error if there is any.
func (c *secrets) Get(name string, options v1.GetOptions) (result *v1alpha1.Secret, err error) {
	result = &v1alpha1.Secret{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("secrets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Secrets that match those selectors.
func (c *secrets) List(opts v1.ListOptions) (result *v1alpha1.SecretList, err error) {
	result = &v1alpha1.SecretList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("secrets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Create takes the representation of a secret and creates it.  Returns the server's representation of the secret, and an error, if there is any.
func (c *secrets) Create(secret *v1alpha1.Secret) (result *v1alpha1.Secret, err error) {
	result = &v1alpha1.Secret{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("secrets").
		Body(secret).
		Do().
		Into(result)
	return
}

// Update takes the representation of a secret and updates it. Returns the server's representation of the secret, and an error, if there is any.
func (c *secrets) Update(secret *v1alpha1.Secret) (result *v1alpha1.Secret, err error) {
	result = &v1alpha1.Secret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("secrets").
		Name(secret.Name).
		Body(secret).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *secrets) UpdateStatus(secret *v1alpha1.Secret) (result *v1alpha1.Secret, err error) {
	result = &v1alpha1.Secret{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("secrets").
		Name(secret.Name).
		SubResource("status").
		Body(secret).
		Do().
		Into(result)
	return
}

// Delete takes name of the secret and deletes it. Returns an error if one occurs.
func (c *secrets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("secrets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *secrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("secrets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched secret.
func (c *secrets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Secret, err error) {
	result = &v1alpha1.Secret{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("secrets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
