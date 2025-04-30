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

	// Addons contains information about enabled/disabled addons and their configuration.
	// +optional
	Addons *gardenercorev1beta1.Addons `json:"addons,omitempty" protobuf:"bytes,1,opt,name=addons"`
	// CloudProfileName is a name of a CloudProfile object.
	// Deprecated: This field will be removed in a future version of Gardener. Use `CloudProfile` instead.
	// Until removed, this field is synced with the `CloudProfile` field.
	// +optional
	CloudProfileName *string `json:"cloudProfileName,omitempty" protobuf:"bytes,2,opt,name=cloudProfileName"`
	// DNS contains information about the DNS settings of the Shoot.
	// +optional
	DNS *gardenercorev1beta1.DNS `json:"dns,omitempty" protobuf:"bytes,3,opt,name=dns"`
	// Extensions contain type and provider information for Shoot extensions.
	// +optional
	Extensions []gardenercorev1beta1.Extension `json:"extensions,omitempty" protobuf:"bytes,4,rep,name=extensions"`
	// Hibernation contains information whether the Shoot is suspended or not.
	// +optional
	Hibernation *gardenercorev1beta1.Hibernation `json:"hibernation,omitempty" protobuf:"bytes,5,opt,name=hibernation"`
	// Kubernetes contains the version and configuration settings of the control plane components.
	Kubernetes gardenercorev1beta1.Kubernetes `json:"kubernetes" protobuf:"bytes,6,opt,name=kubernetes"`
	// Networking contains information about cluster networking such as CNI Plugin type, CIDRs, ...etc.
	// +optional
	Networking *gardenercorev1beta1.Networking `json:"networking,omitempty" protobuf:"bytes,7,opt,name=networking"`
	// Maintenance contains information about the time window for maintenance operations and which
	// operations should be performed.
	// +optional
	Maintenance *gardenercorev1beta1.Maintenance `json:"maintenance,omitempty" protobuf:"bytes,8,opt,name=maintenance"`
	// Monitoring contains information about custom monitoring configurations for the shoot.
	// +optional
	Monitoring *gardenercorev1beta1.Monitoring `json:"monitoring,omitempty" protobuf:"bytes,9,opt,name=monitoring"`
	// Provider contains all provider-specific and provider-relevant information.
	Provider gardenercorev1beta1.Provider `json:"provider" protobuf:"bytes,10,opt,name=provider"`
	// Purpose is the purpose class for this cluster.
	// +optional
	Purpose *gardenercorev1beta1.ShootPurpose `json:"purpose,omitempty" protobuf:"bytes,11,opt,name=purpose,casttype=ShootPurpose"`
	// Region is a name of a region. This field is immutable.
	Region string `json:"region" protobuf:"bytes,12,opt,name=region"`
	// SecretBindingName is the name of a SecretBinding that has a reference to the provider secret.
	// The credentials inside the provider secret will be used to create the shoot in the respective account.
	// The field is mutually exclusive with CredentialsBindingName.
	// This field is immutable.
	// +optional
	SecretBindingName *string `json:"secretBindingName,omitempty" protobuf:"bytes,13,opt,name=secretBindingName"`
	// SeedName is the name of the seed cluster that runs the control plane of the Shoot.
	// +optional
	SeedName *string `json:"seedName,omitempty" protobuf:"bytes,14,opt,name=seedName"`
	// SeedSelector is an optional selector which must match a seed's labels for the shoot to be scheduled on that seed.
	// +optional
	SeedSelector *gardenercorev1beta1.SeedSelector `json:"seedSelector,omitempty" protobuf:"bytes,15,opt,name=seedSelector"`
	// Resources holds a list of named resource references that can be referred to in extension configs by their names.
	// +optional
	Resources []gardenercorev1beta1.NamedResourceReference `json:"resources,omitempty" protobuf:"bytes,16,rep,name=resources"`
	// Tolerations contains the tolerations for taints on seed clusters.
	// +patchMergeKey=key
	// +patchStrategy=merge
	// +optional
	Tolerations []gardenercorev1beta1.Toleration `json:"tolerations,omitempty" patchStrategy:"merge" patchMergeKey:"key" protobuf:"bytes,17,rep,name=tolerations"`
	// ExposureClassName is the optional name of an exposure class to apply a control plane endpoint exposure strategy.
	// This field is immutable.
	// +optional
	ExposureClassName *string `json:"exposureClassName,omitempty" protobuf:"bytes,18,opt,name=exposureClassName"`
	// SystemComponents contains the settings of system components in the control or data plane of the Shoot cluster.
	// +optional
	SystemComponents *gardenercorev1beta1.SystemComponents `json:"systemComponents,omitempty" protobuf:"bytes,19,opt,name=systemComponents"`
	// ControlPlane contains general settings for the control plane of the shoot.
	// +optional
	ControlPlane *gardenercorev1beta1.ControlPlane `json:"controlPlane,omitempty" protobuf:"bytes,20,opt,name=controlPlane"`
	// SchedulerName is the name of the responsible scheduler which schedules the shoot.
	// If not specified, the default scheduler takes over.
	// This field is immutable.
	// +optional
	SchedulerName *string `json:"schedulerName,omitempty" protobuf:"bytes,21,opt,name=schedulerName"`
	// CredentialsBindingName is the name of a CredentialsBinding that has a reference to the provider credentials.
	// The credentials will be used to create the shoot in the respective account. The field is mutually exclusive with SecretBindingName.
	// +optional
	CredentialsBindingName *string `json:"credentialsBindingName,omitempty" protobuf:"bytes,23,opt,name=credentialsBindingName"`
	// AccessRestrictions describe a list of access restrictions for this shoot cluster.
	// +optional
	AccessRestrictions []gardenercorev1beta1.AccessRestrictionWithOptions `json:"accessRestrictions,omitempty" protobuf:"bytes,24,rep,name=accessRestrictions"`
}

// GardenerShootControlPlaneStatus defines the observed state of GardenerShootControlPlane.
type GardenerShootControlPlaneStatus struct {
	// ShootStatus is the status of the Shoot cluster.
	// +optional
	ShootStatus gardenercorev1beta1.ShootStatus `json:"shootStatus"`

	// Initialized denotes that the foo Gardener Shoot control plane API Server is initialized and thus
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
