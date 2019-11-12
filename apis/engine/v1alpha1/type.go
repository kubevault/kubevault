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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleRef contains information that points to the role being used
type RoleRef struct {
	// APIGroup is the group for the resource being referenced
	APIGroup string `json:"apiGroup,omitempty" protobuf:"bytes,1,opt,name=apiGroup"`
	// Kind is the type of resource being referenced
	Kind string `json:"kind,omitempty" protobuf:"bytes,2,opt,name=kind"`
	// Name is the name of resource being referenced
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
	// Namespace is the namespace of the resource being referenced
	Namespace string `json:"namespace" protobuf:"bytes,4,opt,name=namespace"`
}

type RequestConditionType string

// These are the possible conditions for a certificate request.
const (
	AccessApproved RequestConditionType = "Approved"
	AccessDenied   RequestConditionType = "Denied"
)

// Lease contains lease info
type Lease struct {
	// lease id
	ID string `json:"id,omitempty" protobuf:"bytes,1,opt,name=id"`

	// lease duration
	Duration metav1.Duration `json:"duration,omitempty" protobuf:"bytes,2,opt,name=duration"`

	// Specifies whether this lease is renewable
	Renewable bool `json:"renewable,omitempty" protobuf:"varint,3,opt,name=renewable"`
}
