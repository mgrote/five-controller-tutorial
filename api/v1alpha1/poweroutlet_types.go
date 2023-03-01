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
)

const (
	PowerOutletFinalizer = "finalizer.poweroutlets.personal-iot.frup.org"
)

// PoweroutletSpec defines the desired state of Poweroutlet
type PoweroutletSpec struct {
	// The desired switch status.
	// +kubebuilder:default:=OFF
	// +kubebuilder:validation:Enum:=ON;OFF
	Switch           string `json:"switch,omitempty"`
	OutletName       string `json:"outletName,omitempty"`
	MQTTStatusTopik  string `json:"mqttstatustopik,omitempty"`
	MQTTCommandTopik string `json:"mqttcommandtopik,omitempty"`
}

// PoweroutletStatus defines the observed state of Poweroutlet
type PoweroutletStatus struct {
	Switch          string `json:"on,omitempty"`
	Consumption     int32  `json:"consumption,omitempty"`
	ConsumptionUnit string `json:"consumptionunit,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=poweroutlets,scope=Namespaced,categories=all;power,shortName=outlet

// Poweroutlet is the Schema for the poweroutlets API
type Poweroutlet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PoweroutletSpec   `json:"spec,omitempty"`
	Status PoweroutletStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PoweroutletList contains a list of Poweroutlet
type PoweroutletList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Poweroutlet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Poweroutlet{}, &PoweroutletList{})
}
