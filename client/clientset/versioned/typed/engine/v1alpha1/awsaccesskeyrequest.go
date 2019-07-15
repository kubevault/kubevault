/*
Copyright 2019 The Kube Vault Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "kubevault.dev/operator/apis/engine/v1alpha1"
	scheme "kubevault.dev/operator/client/clientset/versioned/scheme"
)

// AWSAccessKeyRequestsGetter has a method to return a AWSAccessKeyRequestInterface.
// A group's client should implement this interface.
type AWSAccessKeyRequestsGetter interface {
	AWSAccessKeyRequests(namespace string) AWSAccessKeyRequestInterface
}

// AWSAccessKeyRequestInterface has methods to work with AWSAccessKeyRequest resources.
type AWSAccessKeyRequestInterface interface {
	Create(*v1alpha1.AWSAccessKeyRequest) (*v1alpha1.AWSAccessKeyRequest, error)
	Update(*v1alpha1.AWSAccessKeyRequest) (*v1alpha1.AWSAccessKeyRequest, error)
	UpdateStatus(*v1alpha1.AWSAccessKeyRequest) (*v1alpha1.AWSAccessKeyRequest, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.AWSAccessKeyRequest, error)
	List(opts v1.ListOptions) (*v1alpha1.AWSAccessKeyRequestList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSAccessKeyRequest, err error)
	AWSAccessKeyRequestExpansion
}

// aWSAccessKeyRequests implements AWSAccessKeyRequestInterface
type aWSAccessKeyRequests struct {
	client rest.Interface
	ns     string
}

// newAWSAccessKeyRequests returns a AWSAccessKeyRequests
func newAWSAccessKeyRequests(c *EngineV1alpha1Client, namespace string) *aWSAccessKeyRequests {
	return &aWSAccessKeyRequests{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the aWSAccessKeyRequest, and returns the corresponding aWSAccessKeyRequest object, and an error if there is any.
func (c *aWSAccessKeyRequests) Get(name string, options v1.GetOptions) (result *v1alpha1.AWSAccessKeyRequest, err error) {
	result = &v1alpha1.AWSAccessKeyRequest{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AWSAccessKeyRequests that match those selectors.
func (c *aWSAccessKeyRequests) List(opts v1.ListOptions) (result *v1alpha1.AWSAccessKeyRequestList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.AWSAccessKeyRequestList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested aWSAccessKeyRequests.
func (c *aWSAccessKeyRequests) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a aWSAccessKeyRequest and creates it.  Returns the server's representation of the aWSAccessKeyRequest, and an error, if there is any.
func (c *aWSAccessKeyRequests) Create(aWSAccessKeyRequest *v1alpha1.AWSAccessKeyRequest) (result *v1alpha1.AWSAccessKeyRequest, err error) {
	result = &v1alpha1.AWSAccessKeyRequest{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		Body(aWSAccessKeyRequest).
		Do().
		Into(result)
	return
}

// Update takes the representation of a aWSAccessKeyRequest and updates it. Returns the server's representation of the aWSAccessKeyRequest, and an error, if there is any.
func (c *aWSAccessKeyRequests) Update(aWSAccessKeyRequest *v1alpha1.AWSAccessKeyRequest) (result *v1alpha1.AWSAccessKeyRequest, err error) {
	result = &v1alpha1.AWSAccessKeyRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		Name(aWSAccessKeyRequest.Name).
		Body(aWSAccessKeyRequest).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *aWSAccessKeyRequests) UpdateStatus(aWSAccessKeyRequest *v1alpha1.AWSAccessKeyRequest) (result *v1alpha1.AWSAccessKeyRequest, err error) {
	result = &v1alpha1.AWSAccessKeyRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		Name(aWSAccessKeyRequest.Name).
		SubResource("status").
		Body(aWSAccessKeyRequest).
		Do().
		Into(result)
	return
}

// Delete takes name of the aWSAccessKeyRequest and deletes it. Returns an error if one occurs.
func (c *aWSAccessKeyRequests) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *aWSAccessKeyRequests) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched aWSAccessKeyRequest.
func (c *aWSAccessKeyRequests) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AWSAccessKeyRequest, err error) {
	result = &v1alpha1.AWSAccessKeyRequest{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("awsaccesskeyrequests").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
