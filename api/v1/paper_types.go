/*
Copyright 2022 Bernhard Aichinger.

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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PaperSpec defines the desired state of Paper
type PaperSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=\d+.\d+.\d+
	Version string `json:"version"`
}

// PaperStatus defines the observed state of Paper
type PaperStatus struct {
	Conditions       []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	DesiredState     *DesiredState      `json:"desiredState,omitempty"`
	ActualState      *ActualState       `json:"actualState,omitempty"`
	UpdatedTimestamp *metav1.Time       `json:"updatedTimestamp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Paper is the Schema for the papers API
type Paper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec   PaperSpec   `json:"spec"`
	Status PaperStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PaperList contains a list of Paper
type PaperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Paper `json:"items"`
}

type Version struct {
	Version string `json:"version,omitempty"`
	Build   int    `json:"build,omitempty"`
}

type DesiredState struct {
	Version          Version     `json:"version,omitempty"`
	Url              string      `json:"url,omitempty"`
	UpdatedTimestamp metav1.Time `json:"updatedTimestamp,omitempty"`
}

type ActualState struct {
	Version Version `json:"version,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Paper{}, &PaperList{})
}

func (dv *Version) String() string {
	return fmt.Sprintf("%s-%d", strings.Replace(dv.Version, ".", "-", -1), dv.Build)
}
