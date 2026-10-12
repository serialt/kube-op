/*
Copyright 2026.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CaddySpec defines the desired state of Caddy.
type CaddySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Replicas int32  `json:"replicas"`         // 对应 YAML 的 `spec.replicas`
	Image    string `json:"image"`            // 对应 YAML 的 `spec.image`
	Port     int32  `json:"port"`             // 应用监听的端口
	SvcPort  int32  `json:"svcPort"`          // Service 暴露的端口
	Domain   string `json:"domain,omitempty"` // Ingress 暴露使用的域名；未设置时不创建 Ingress
}

// CaddyStatus defines the observed state of Caddy.
type CaddyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
	AvailableReplicas int32              `json:"available_replicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Caddy is the Schema for the caddies API.
type Caddy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaddySpec   `json:"spec,omitempty"`
	Status CaddyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CaddyList contains a list of Caddy.
type CaddyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Caddy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Caddy{}, &CaddyList{})
}
