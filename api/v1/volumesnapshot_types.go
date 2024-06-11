/*
Copyright 2024.

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

// VolumeSnapshotSpec defines the desired state of VolumeSnapshot
type VolumeSnapshotSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	PvcSelectorLabel string `json:"pvcselectorlabel,omitempty"`
}

// VolumeSnapshotStatus defines the observed state of VolumeSnapshot
type VolumeSnapshotStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`

// VolumeSnapshot is the Schema for the volumesnapshots API
type VolumeSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeSnapshotSpec   `json:"spec,omitempty"`
	Status VolumeSnapshotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VolumeSnapshotList contains a list of VolumeSnapshot
type VolumeSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeSnapshot{}, &VolumeSnapshotList{})
}
