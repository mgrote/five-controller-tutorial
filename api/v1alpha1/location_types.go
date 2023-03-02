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
	LocationFinalizer    = "finalizer.locations.personal-iot.frup.org"
	LocationMoodDark     = "DARK"
	LocationMoodBright   = "BRIGHT"
	LocationMoodDontKnow = "DONTKNOW"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LocationSpec defines the desired state of Location
type LocationSpec struct {
	// The mood the location should be in.
	// +kubebuilder:default:=DARK
	// +kubebuilder:validation:Enum:=DARK;BRIGHT;DONTKNOW
	Mood string `json:"mood,omitempty"`
}

// LocationStatus defines the observed state of Location
type LocationStatus struct {
	// The mood the location currently is.
	Mood            string `json:"mood"`
	Consumption     int32  `json:"consumption,omitempty"`
	ConsumptionUnit string `json:"consumptionunit,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=locations,scope=Namespaced,categories=all,shortName=loc

// Location is the Schema for the locations API
type Location struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocationSpec   `json:"spec,omitempty"`
	Status LocationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LocationList contains a list of Location
type LocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Location `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Location{}, &LocationList{})
}
