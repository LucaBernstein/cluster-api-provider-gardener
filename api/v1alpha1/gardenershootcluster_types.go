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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GardenerShootCluster represents a Shoot cluster.
type GardenerShootCluster struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the Shoot cluster.
	// If the object's deletion timestamp is set, this field is immutable.
	// +optional
	Spec   gardenercorev1beta1.ShootSpec `json:"spec,omitempty"`
	Status GardenerShootClusterStatus    `json:"status,omitempty"`
}

// GardenerShootClusterStatus defines the observed state of GardenerShootCluster.
type GardenerShootClusterStatus struct {
	// Ready denotes that the foo cluster infrastructure is fully provisioned.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the infra cluster.
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true

// GardenerShootClusterList contains a list of GardenerShootCluster.
type GardenerShootClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GardenerShootCluster `json:"items"`
}

func init() {
	objectTypes = append(objectTypes, &GardenerShootCluster{}, &GardenerShootClusterList{})
}
