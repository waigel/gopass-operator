/*
Copyright 2022.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GopassItemSpec defines the desired state of GopassItem
type GopassItemSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of GopassItem. Edit gopassitem_types.go to remove/update
	Foo      string `json:"foo,omitempty"`
	ItemPath string `json:"itemPath,omitempty"`
}

// GopassItemStatus defines the observed state of GopassItem
type GopassItemStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GopassItem is the Schema for the gopassitems API
type GopassItem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Kubernetes secret type. More info: https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
	Type string `json:"type,omitempty"`

	Spec   GopassItemSpec   `json:"spec,omitempty"`
	Status GopassItemStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GopassItemList contains a list of GopassItem
type GopassItemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GopassItem `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GopassItem{}, &GopassItemList{})
}
