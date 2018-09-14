/*
Copyright 2018 The Kube Vault Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kubevault/operator/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// VaultserverVersionLister helps list VaultserverVersions.
type VaultserverVersionLister interface {
	// List lists all VaultserverVersions in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.VaultserverVersion, err error)
	// Get retrieves the VaultserverVersion from the index for a given name.
	Get(name string) (*v1alpha1.VaultserverVersion, error)
	VaultserverVersionListerExpansion
}

// vaultserverVersionLister implements the VaultserverVersionLister interface.
type vaultserverVersionLister struct {
	indexer cache.Indexer
}

// NewVaultserverVersionLister returns a new VaultserverVersionLister.
func NewVaultserverVersionLister(indexer cache.Indexer) VaultserverVersionLister {
	return &vaultserverVersionLister{indexer: indexer}
}

// List lists all VaultserverVersions in the indexer.
func (s *vaultserverVersionLister) List(selector labels.Selector) (ret []*v1alpha1.VaultserverVersion, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.VaultserverVersion))
	})
	return ret, err
}

// Get retrieves the VaultserverVersion from the index for a given name.
func (s *vaultserverVersionLister) Get(name string) (*v1alpha1.VaultserverVersion, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("vaultserverversion"), name)
	}
	return obj.(*v1alpha1.VaultserverVersion), nil
}
