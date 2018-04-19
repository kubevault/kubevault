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

package fake

import (
	v1alpha1 "github.com/soter/vault-operator/apis/extensions/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	testing "k8s.io/client-go/testing"
)

// FakeSnapshots implements SnapshotInterface
type FakeSnapshots struct {
	Fake *FakeExtensionsV1alpha1
	ns   string
}

var snapshotsResource = schema.GroupVersionResource{Group: "extensions.vault.soter.ac", Version: "v1alpha1", Resource: "snapshots"}

var snapshotsKind = schema.GroupVersionKind{Group: "extensions.vault.soter.ac", Version: "v1alpha1", Kind: "Snapshot"}

// Get takes name of the snapshot, and returns the corresponding snapshot object, and an error if there is any.
func (c *FakeSnapshots) Get(name string, options v1.GetOptions) (result *v1alpha1.Snapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(snapshotsResource, c.ns, name), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}

// List takes label and field selectors, and returns the list of Snapshots that match those selectors.
func (c *FakeSnapshots) List(opts v1.ListOptions) (result *v1alpha1.SnapshotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(snapshotsResource, snapshotsKind, c.ns, opts), &v1alpha1.SnapshotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SnapshotList{}
	for _, item := range obj.(*v1alpha1.SnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSnapshots) UpdateStatus(snapshot *v1alpha1.Snapshot) (*v1alpha1.Snapshot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(snapshotsResource, "status", c.ns, snapshot), &v1alpha1.Snapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Snapshot), err
}

// Delete takes name of the snapshot and deletes it. Returns an error if one occurs.
func (c *FakeSnapshots) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(snapshotsResource, c.ns, name), &v1alpha1.Snapshot{})

	return err
}
