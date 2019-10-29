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
package inmem

import (
	core "k8s.io/api/core/v1"
)

var inmenStorage = `
storage "inmem" {
}
`

type Options struct{}

func NewOptions() (*Options, error) {
	return &Options{}, nil
}

func (o *Options) Apply(pt *core.PodTemplateSpec) error {
	return nil
}

// GetStorageConfig will create storage config for inmem backend
func (o *Options) GetStorageConfig() (string, error) {
	return inmenStorage, nil
}
