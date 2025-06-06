/*
Copyright 2025.

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
	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GSCReferenceNamespaceKey     = "infrastructure.cluster.x-k8s.io/gsc_namespace"
	GSCReferenceNameKey          = "infrastructure.cluster.x-k8s.io/gsc_name"
	GSCReferecenceClusterNameKey = "controlplane.cluster.x-k8s.io/gsc_cluster"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GardenerShootClusterSpec defines the desired state of GardenerShootCluster.
type GardenerShootClusterSpec struct {
	// Hibernation contains information whether the Shoot is suspended or not.
	// +optional
	Hibernation *gardenercorev1beta1.Hibernation `json:"hibernation,omitempty" protobuf:"bytes,5,opt,name=hibernation"`
	// Maintenance contains information about the time window for maintenance operations and which
	// operations should be performed.
	// +optional
	Maintenance *gardenercorev1beta1.Maintenance `json:"maintenance,omitempty" protobuf:"bytes,8,opt,name=maintenance"`
	// Region is a name of a region. This field is immutable.
	Region string `json:"region" protobuf:"bytes,12,opt,name=region"`
	// SeedName is the name of the seed cluster that runs the control plane of the Shoot.
	// +optional
	SeedName *string `json:"seedName,omitempty" protobuf:"bytes,14,opt,name=seedName"`
	// SeedSelector is an optional selector which must match a seed's labels for the shoot to be scheduled on that seed.
	// +optional
	SeedSelector *gardenercorev1beta1.SeedSelector `json:"seedSelector,omitempty" protobuf:"bytes,15,opt,name=seedSelector"`
}

// GardenerShootClusterStatus defines the observed state of GardenerShootCluster.
type GardenerShootClusterStatus struct {
	// Ready denotes that the Seed where the Shoot is hosted is ready.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infa cluster.
	// +optional
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// GardenerShootCluster is the Schema for the gardenershootclusters API.
type GardenerShootCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GardenerShootClusterSpec   `json:"spec,omitempty"`
	Status GardenerShootClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GardenerShootClusterList contains a list of GardenerShootCluster.
type GardenerShootClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GardenerShootCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GardenerShootCluster{}, &GardenerShootClusterList{})
}
