/*
Copyright 2023 mgrote.

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
	controllerruntime "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type MQTTConfig struct {
	Broker   *string `json:"broker,omitempty"`
	ClientID *string `json:"clientID,omitempty"`
	// TODO should be a secret
	UserName *string `json:"userName,omitempty"`
	Password *string `json:"password,omitempty"`
}

type MQTTProjectConfigSpec struct {
	// empty
}

type MQTTControllerConfigStatus struct {
	// empty
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MQTTControllerConfig is the Schema for the mqttcontrollerconfigs API
type MQTTControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	controllerruntime.ControllerManagerConfigurationSpec `json:",inline"`

	Spec   MQTTProjectConfigSpec      `json:"spec,omitempty"`
	Status MQTTControllerConfigStatus `json:"status,omitempty"`

	MQTTConfig MQTTConfig `json:"mqttConfig,omitempty"`
}

//+kubebuilder:object:root=true

// MQTTControllerConfigList contains a list of MQTTControllerConfig
type MQTTControllerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MQTTControllerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MQTTControllerConfig{}, &MQTTControllerConfigList{})
}
