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
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	GSCPReferenceNamespaceKey     = "controlplane.cluster.x-k8s.io/gscp_namespace"
	GSCPReferenceNameKey          = "controlplane.cluster.x-k8s.io/gscp_name"
	GSCPReferecenceClusterNameKey = "controlplane.cluster.x-k8s.io/gscp_cluster"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gscp
// +kubebuilder:printcolumn:name="Initialized",type=boolean,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// GardenerShootControlPlane represents a Shoot cluster.
type GardenerShootControlPlane struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the Shoot cluster.
	// If the object's deletion timestamp is set, this field is immutable.
	// +optional
	Spec   GardenerShootControlPlaneSpec   `json:"spec,omitempty"`
	Status GardenerShootControlPlaneStatus `json:"status,omitempty"`
}

// GardenerShootControlPlaneSpec represents the Spec of the Shoot Cluster,
// as well as the fields defined by the Cluster API contract.
type GardenerShootControlPlaneSpec struct {
	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint clusterv1beta1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// Version defines the desired Kubernetes version for the control plane.
	// The value must be a valid semantic version; also if the value provided by the user does not start with the v prefix, it
	// must be added.
	// +optional
	Version string `json:"version,omitempty"`

	// ProjectNamespace is the namespace in which the Shoot should be placed in.
	// This has to be a valid project namespace within the Gardener cluster.
	// If not set, the namespace of this object will be used in the Gardener cluster.
	// +optional
	ProjectNamespace string `json:"projectNamespace,omitempty"`

	// ShootSpec is the specification of the desired Shoot cluster.
	// + optional
	ShootSpec gardenercorev1beta1.ShootSpec `json:"shootSpec,omitempty"`
}

// GardenerShootControlPlaneStatus defines the observed state of GardenerShootControlPlane.
type GardenerShootControlPlaneStatus struct {
	// ShootStatus is the status of the Shoot cluster.
	// +optional
	ShootStatus gardenercorev1beta1.ShootStatus `json:"shootStatus"`

	// Initialized denotes that the foo control plane  API Server is initialized and thus
	// it can accept requests.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the control plane.
	// +optional
	// +kubebuilder:default=false
	Initialized bool `json:"initialized"`

	// Ready denotes that the foo control plane is ready to serve requests.
	// NOTE: this field is part of the Cluster API contract and it is used to orchestrate provisioning.
	// The value of this field is never updated after provisioning is completed. Please use conditions
	// to check the operational state of the control plane.
	// +optional
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true

// GardenerShootControlPlaneList contains a list of GardenerShootControlPlane.
type GardenerShootControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GardenerShootControlPlane `json:"items"`
}

func init() {
	objectTypes = append(objectTypes, &GardenerShootControlPlane{}, &GardenerShootControlPlaneList{})
}
