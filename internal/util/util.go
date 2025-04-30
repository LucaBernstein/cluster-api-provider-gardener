package util

import (
	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	controlplanev1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/controlplane/v1alpha1"
	infrastructurev1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/infrastructure/v1alpha1"
)

func ShootNameFromCAPIResources(cluster clusterv1beta1.Cluster, controlPlane controlplanev1alpha1.GardenerShootControlPlane) types.NamespacedName {
	return types.NamespacedName{
		Name:      cluster.Name,
		Namespace: controlPlane.Spec.ProjectNamespace,
	}
}

func ShootFromCAPIResources(
	capiCluster clusterv1beta1.Cluster,
	controlPlane controlplanev1alpha1.GardenerShootControlPlane,
	infraCluster infrastructurev1alpha1.GardenerShootCluster,
) *gardenercorev1beta1.Shoot {
	namespacedName := ShootNameFromCAPIResources(capiCluster, controlPlane)

	infraSpec := infraCluster.Spec
	controlPlaneSpec := controlPlane.Spec

	return &gardenercorev1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: gardenercorev1beta1.ShootSpec{
			Addons:                 controlPlaneSpec.Addons,
			DNS:                    controlPlaneSpec.DNS,
			Extensions:             controlPlaneSpec.Extensions,
			Hibernation:            controlPlaneSpec.Hibernation,
			Kubernetes:             controlPlaneSpec.Kubernetes,
			Networking:             controlPlaneSpec.Networking,
			Maintenance:            controlPlaneSpec.Maintenance,
			Monitoring:             controlPlaneSpec.Monitoring,
			Provider:               controlPlaneSpec.Provider,
			Purpose:                controlPlaneSpec.Purpose,
			Region:                 controlPlaneSpec.Region,
			SecretBindingName:      controlPlaneSpec.SecretBindingName,
			SeedName:               controlPlaneSpec.SeedName,
			SeedSelector:           controlPlaneSpec.SeedSelector,
			Resources:              controlPlaneSpec.Resources,
			Tolerations:            controlPlaneSpec.Tolerations,
			ExposureClassName:      controlPlaneSpec.ExposureClassName,
			SystemComponents:       controlPlaneSpec.SystemComponents,
			ControlPlane:           controlPlaneSpec.ControlPlane,
			SchedulerName:          controlPlaneSpec.SchedulerName,
			CloudProfile:           infraSpec.CloudProfile,
			CredentialsBindingName: controlPlaneSpec.CredentialsBindingName,
			AccessRestrictions:     controlPlaneSpec.AccessRestrictions,
		},
	}
}

func SyncShootSpecFromGSCP(shoot *gardenercorev1beta1.Shoot, controlPlane *controlplanev1alpha1.GardenerShootControlPlane) {
	controlPlaneSpec := controlPlane.Spec

	shoot.Spec.Addons = controlPlaneSpec.Addons
	shoot.Spec.DNS = controlPlaneSpec.DNS
	shoot.Spec.Extensions = controlPlaneSpec.Extensions
	shoot.Spec.Hibernation = controlPlaneSpec.Hibernation
	shoot.Spec.Kubernetes = controlPlaneSpec.Kubernetes
	shoot.Spec.Networking = controlPlaneSpec.Networking
	shoot.Spec.Maintenance = controlPlaneSpec.Maintenance
	shoot.Spec.Monitoring = controlPlaneSpec.Monitoring
	shoot.Spec.Provider = controlPlaneSpec.Provider
	shoot.Spec.Purpose = controlPlaneSpec.Purpose
	shoot.Spec.Region = controlPlaneSpec.Region
	shoot.Spec.SecretBindingName = controlPlaneSpec.SecretBindingName
	// Let's not allow updates on SeedName as this causes the reconciler to not be able to update anything
	// shoot.Spec.SeedName = controlPlaneSpec.SeedName
	shoot.Spec.SeedSelector = controlPlaneSpec.SeedSelector
	shoot.Spec.Resources = controlPlaneSpec.Resources
	shoot.Spec.Tolerations = controlPlaneSpec.Tolerations
	shoot.Spec.ExposureClassName = controlPlaneSpec.ExposureClassName
	shoot.Spec.SystemComponents = controlPlaneSpec.SystemComponents
	shoot.Spec.ControlPlane = controlPlaneSpec.ControlPlane
	shoot.Spec.SchedulerName = controlPlaneSpec.SchedulerName
}

func SyncGSCPSpecFromShoot(shoot *gardenercorev1beta1.Shoot, controlPlane *controlplanev1alpha1.GardenerShootControlPlane) {
	controlPlane.Spec.Addons = shoot.Spec.Addons
	controlPlane.Spec.DNS = shoot.Spec.DNS
	controlPlane.Spec.Extensions = shoot.Spec.Extensions
	controlPlane.Spec.Hibernation = shoot.Spec.Hibernation
	controlPlane.Spec.Kubernetes = shoot.Spec.Kubernetes
	controlPlane.Spec.Networking = shoot.Spec.Networking
	controlPlane.Spec.Maintenance = shoot.Spec.Maintenance
	controlPlane.Spec.Monitoring = shoot.Spec.Monitoring
	controlPlane.Spec.Provider = shoot.Spec.Provider
	controlPlane.Spec.Purpose = shoot.Spec.Purpose
	controlPlane.Spec.Region = shoot.Spec.Region
	controlPlane.Spec.SecretBindingName = shoot.Spec.SecretBindingName
	// Let's not allow updates on SeedName as this causes the reconciler to not be able to update anything
	// shoot.Spec.SeedName = controlPlaneSpec.SeedName
	controlPlane.Spec.SeedSelector = shoot.Spec.SeedSelector
	controlPlane.Spec.Resources = shoot.Spec.Resources
	controlPlane.Spec.Tolerations = shoot.Spec.Tolerations
	controlPlane.Spec.ExposureClassName = shoot.Spec.ExposureClassName
	controlPlane.Spec.SystemComponents = shoot.Spec.SystemComponents
	controlPlane.Spec.ControlPlane = shoot.Spec.ControlPlane
	controlPlane.Spec.SchedulerName = shoot.Spec.SchedulerName
	controlPlane.Spec.CredentialsBindingName = shoot.Spec.CredentialsBindingName
	controlPlane.Spec.AccessRestrictions = shoot.Spec.AccessRestrictions
}

func SyncShootSpecFromCluster(shoot *gardenercorev1beta1.Shoot, cluster *infrastructurev1alpha1.GardenerShootCluster) {
	shoot.Spec.CloudProfile = cluster.Spec.CloudProfile
}

func SyncClusterSpecFromShoot(shoot *gardenercorev1beta1.Shoot, cluster *infrastructurev1alpha1.GardenerShootCluster) {
	cluster.Spec.CloudProfile = shoot.Spec.CloudProfile
}
