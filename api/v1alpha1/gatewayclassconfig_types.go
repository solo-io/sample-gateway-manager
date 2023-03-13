/*
Copyright 2023.

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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=gateway-api,shortName=gcc

// GatewayClassConfig is the Schema for the gatewayclassconfigs API.
type GatewayClassConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewayClassConfigSpec   `json:"spec,omitempty"`
	Status GatewayClassConfigStatus `json:"status,omitempty"`
}

// GatewayClassConfigSpec defines the desired state of GatewayClassConfig.
type GatewayClassConfigSpec struct {
	// Foo is an example field that represents Gateway configuration.
	//
	// If unset, defaults to "bar".
	//
	// +kubebuilder:default="bar"
	Foo string `json:"foo"`
}

// GatewayClassConfigStatus defines the observed state of GatewayClassConfig.
type GatewayClassConfigStatus struct {
	// ObservedFoo is an example status field that is set when the GatewayClassConfig
	// is reconciled.
	//
	ObservedFoo string `json:"observedFoo"`

	// Conditions represent the observation state of the GatewayClassConfig.
	//
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true

// GatewayClassConfigList contains a list of GatewayClassConfig
type GatewayClassConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayClassConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GatewayClassConfig{}, &GatewayClassConfigList{})
}
