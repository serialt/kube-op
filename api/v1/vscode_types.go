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

// VSCodeSpec defines the desired state of VSCode.
type VSCodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image            string `json:"image"`                      // 对应 YAML 的 `spec.image`
	Port             int32  `json:"port"`                       // 应用监听的端口
	SvcPort          int32  `json:"svcPort"`                    // Service 暴露的端口
	Domain           string `json:"domain,omitempty"`           // Ingress 暴露使用的域名；未设置时不创建 Ingress
	IngressClassName string `json:"ingressClassName,omitempty"` // IngressClassName 用于指定 Ingress 的类名；未设置时使用默认类IngressClassName
}

// VSCodeStatus defines the observed state of VSCode.
type VSCodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions        []metav1.Condition `json:"conditions,omitempty"`
	AvailableReplicas int32              `json:"available_replicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VSCode is the Schema for the vscodes API.
type VSCode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSCodeSpec   `json:"spec,omitempty"`
	Status VSCodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VSCodeList contains a list of VSCode.
type VSCodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSCode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSCode{}, &VSCodeList{})
}
