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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindKubeVaultOperator = "KubeVaultOperator"
	ResourceKubeVaultOperator     = "kubevaultoperator"
	ResourceKubeVaultOperators    = "kubevaultoperators"
)

// KubeVaultOperator defines the schama for KubeVault Operator Installer.

// +genclient
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kubevaultoperators,singular=kubevaultoperator,categories={kubevault,appscode}
type KubeVaultOperator struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              KubeVaultOperatorSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type ImageRef struct {
	Registry   string `json:"registry" protobuf:"bytes,1,opt,name=registry"`
	Repository string `json:"repository" protobuf:"bytes,2,opt,name=repository"`
	Tag        string `json:"tag" protobuf:"bytes,3,opt,name=tag"`
}

// KubeVaultOperatorSpec is the spec for redis version
type KubeVaultOperatorSpec struct {
	ReplicaCount    int32             `json:"replicaCount" protobuf:"varint,1,opt,name=replicaCount"`
	KubeVault       ImageRef          `json:"kubevault" protobuf:"bytes,2,opt,name=kubevault"`
	Cleaner         ImageRef          `json:"cleaner" protobuf:"bytes,3,opt,name=cleaner"`
	ImagePullPolicy string            `json:"imagePullPolicy" protobuf:"bytes,4,opt,name=imagePullPolicy"`
	CriticalAddon   bool              `json:"criticalAddon" protobuf:"varint,5,opt,name=criticalAddon"`
	LogLevel        int32             `json:"logLevel" protobuf:"varint,6,opt,name=logLevel"`
	Annotations     map[string]string `json:"annotations" protobuf:"bytes,7,rep,name=annotations"`
	NodeSelector    map[string]string `json:"nodeSelector" protobuf:"bytes,8,rep,name=nodeSelector"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []core.Toleration `json:"tolerations,omitempty" protobuf:"bytes,9,rep,name=tolerations"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity        *core.Affinity     `json:"affinity,omitempty" protobuf:"bytes,10,opt,name=affinity"`
	ServiceAccount  ServiceAccountSpec `json:"serviceAccount" protobuf:"bytes,11,opt,name=serviceAccount"`
	Apiserver       WebHookSpec        `json:"apiserver" protobuf:"bytes,12,opt,name=apiserver"`
	EnableAnalytics bool               `json:"enableAnalytics" protobuf:"varint,13,opt,name=enableAnalytics"`
	Monitoring      Monitoring         `json:"monitoring" protobuf:"bytes,14,opt,name=monitoring"`
	ClusterName     string             `json:"clusterName" protobuf:"bytes,15,opt,name=clusterName"`
}

type ServiceAccountSpec struct {
	Create bool   `json:"create" protobuf:"varint,1,opt,name=create"`
	Name   string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

type WebHookSpec struct {
	GroupPriorityMinimum        int32           `json:"groupPriorityMinimum" protobuf:"varint,1,opt,name=groupPriorityMinimum"`
	VersionPriority             int32           `json:"versionPriority" protobuf:"varint,2,opt,name=versionPriority"`
	EnableMutatingWebhook       bool            `json:"enableMutatingWebhook" protobuf:"varint,3,opt,name=enableMutatingWebhook"`
	EnableValidatingWebhook     bool            `json:"enableValidatingWebhook" protobuf:"varint,4,opt,name=enableValidatingWebhook"`
	Ca                          string          `json:"ca" protobuf:"bytes,5,opt,name=ca"`
	BypassValidatingWebhookXray bool            `json:"bypassValidatingWebhookXray" protobuf:"varint,6,opt,name=bypassValidatingWebhookXray"`
	UseKubeapiserverFqdnForAks  bool            `json:"useKubeapiserverFqdnForAks" protobuf:"varint,7,opt,name=useKubeapiserverFqdnForAks"`
	Healthcheck                 HealthcheckSpec `json:"healthcheck" protobuf:"bytes,8,opt,name=healthcheck"`
}

type HealthcheckSpec struct {
	Enabled bool `json:"enabled" protobuf:"varint,1,opt,name=enabled"`
}

type Monitoring struct {
	Agent          string               `json:"agent" protobuf:"bytes,1,opt,name=agent"`
	Operator       bool                 `json:"operator" protobuf:"varint,2,opt,name=operator"`
	Prometheus     PrometheusSpec       `json:"prometheus" protobuf:"bytes,3,opt,name=prometheus"`
	ServiceMonitor ServiceMonitorLabels `json:"serviceMonitor" protobuf:"bytes,4,opt,name=serviceMonitor"`
}

type PrometheusSpec struct {
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
}

type ServiceMonitorLabels struct {
	Labels map[string]string `json:"labels" protobuf:"bytes,1,rep,name=labels"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubeVaultOperatorList is a list of KubeVaultOperators
type KubeVaultOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of KubeVaultOperator CRD objects
	Items []KubeVaultOperator `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
