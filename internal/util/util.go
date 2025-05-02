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

	return &gardenercorev1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: gardenercorev1beta1.ShootSpec{
			Addons:                 controlPlane.Spec.Addons,
			DNS:                    controlPlane.Spec.DNS,
			Extensions:             controlPlane.Spec.Extensions,
			Hibernation:            infraCluster.Spec.Hibernation,
			Kubernetes:             controlPlane.Spec.Kubernetes,
			Networking:             controlPlane.Spec.Networking,
			Maintenance:            infraCluster.Spec.Maintenance,
			Monitoring:             controlPlane.Spec.Monitoring,
			Provider:               controlPlane.Spec.Provider,
			Purpose:                controlPlane.Spec.Purpose,
			Region:                 infraCluster.Spec.Region,
			SecretBindingName:      controlPlane.Spec.SecretBindingName,
			SeedName:               infraCluster.Spec.SeedName,
			SeedSelector:           infraCluster.Spec.SeedSelector,
			Resources:              controlPlane.Spec.Resources,
			Tolerations:            controlPlane.Spec.Tolerations,
			ExposureClassName:      controlPlane.Spec.ExposureClassName,
			SystemComponents:       controlPlane.Spec.SystemComponents,
			ControlPlane:           controlPlane.Spec.ControlPlane,
			SchedulerName:          controlPlane.Spec.SchedulerName,
			CloudProfile:           controlPlane.Spec.CloudProfile,
			CredentialsBindingName: controlPlane.Spec.CredentialsBindingName,
			AccessRestrictions:     controlPlane.Spec.AccessRestrictions,
		},
	}
}

func SyncShootSpecFromGSCP(shoot *gardenercorev1beta1.Shoot, controlPlane *controlplanev1alpha1.GardenerShootControlPlane) {

	shoot.Spec.Addons = controlPlane.Spec.Addons
	shoot.Spec.DNS = controlPlane.Spec.DNS
	shoot.Spec.Extensions = controlPlane.Spec.Extensions
	shoot.Spec.Kubernetes = controlPlane.Spec.Kubernetes
	shoot.Spec.Networking = controlPlane.Spec.Networking
	shoot.Spec.Monitoring = controlPlane.Spec.Monitoring
	shoot.Spec.Provider = controlPlane.Spec.Provider
	shoot.Spec.Purpose = controlPlane.Spec.Purpose
	shoot.Spec.SecretBindingName = controlPlane.Spec.SecretBindingName
	// Let's not allow updates on SeedName as this causes the reconciler to not be able to update anything
	// shoot.Spec.SeedName = controlPlane.Spec.SeedName
	shoot.Spec.Resources = controlPlane.Spec.Resources
	shoot.Spec.Tolerations = controlPlane.Spec.Tolerations
	shoot.Spec.ExposureClassName = controlPlane.Spec.ExposureClassName
	shoot.Spec.SystemComponents = controlPlane.Spec.SystemComponents
	shoot.Spec.ControlPlane = controlPlane.Spec.ControlPlane
	shoot.Spec.SchedulerName = controlPlane.Spec.SchedulerName
}

func SyncGSCPSpecFromShoot(shoot *gardenercorev1beta1.Shoot, controlPlane *controlplanev1alpha1.GardenerShootControlPlane) {
	controlPlane.Spec.Addons = shoot.Spec.Addons
	controlPlane.Spec.DNS = shoot.Spec.DNS
	controlPlane.Spec.Extensions = shoot.Spec.Extensions
	controlPlane.Spec.Kubernetes = shoot.Spec.Kubernetes
	controlPlane.Spec.Networking = shoot.Spec.Networking
	controlPlane.Spec.Monitoring = shoot.Spec.Monitoring
	controlPlane.Spec.Provider = shoot.Spec.Provider
	controlPlane.Spec.Purpose = shoot.Spec.Purpose
	controlPlane.Spec.SecretBindingName = shoot.Spec.SecretBindingName
	controlPlane.Spec.Resources = shoot.Spec.Resources
	controlPlane.Spec.Tolerations = shoot.Spec.Tolerations
	controlPlane.Spec.ExposureClassName = shoot.Spec.ExposureClassName
	controlPlane.Spec.SystemComponents = shoot.Spec.SystemComponents
	controlPlane.Spec.ControlPlane = shoot.Spec.ControlPlane
	controlPlane.Spec.SchedulerName = shoot.Spec.SchedulerName
	controlPlane.Spec.CredentialsBindingName = shoot.Spec.CredentialsBindingName
	controlPlane.Spec.AccessRestrictions = shoot.Spec.AccessRestrictions
}

func SyncShootSpecFromCluster(shoot *gardenercorev1beta1.Shoot, infraCluster *infrastructurev1alpha1.GardenerShootCluster) {
	shoot.Spec.Hibernation = infraCluster.Spec.Hibernation
	// Do not allow to nil the maintenance field as this will cause a potential eternal reconciliation loop,
	// because the maintenance time window is defaulted to a random time window, which causes problems when syncing.
	if infraCluster.Spec.Hibernation != nil {
		shoot.Spec.Maintenance = infraCluster.Spec.Maintenance
	}
	shoot.Spec.Region = infraCluster.Spec.Region
	shoot.Spec.SeedName = infraCluster.Spec.SeedName
	shoot.Spec.SeedSelector = infraCluster.Spec.SeedSelector
}

func SyncClusterSpecFromShoot(shoot *gardenercorev1beta1.Shoot, infraCluster *infrastructurev1alpha1.GardenerShootCluster) {
	infraCluster.Spec.Hibernation = shoot.Spec.Hibernation
	// Do not allow to nil the maintenance field as this will cause a potential eternal reconciliation loop,
	// because the maintenance time window is defaulted to a random time window, which causes problems when syncing.
	if shoot.Spec.Hibernation != nil {
		infraCluster.Spec.Maintenance = shoot.Spec.Maintenance
	}
	infraCluster.Spec.Region = shoot.Spec.Region
	infraCluster.Spec.SeedName = shoot.Spec.SeedName
	infraCluster.Spec.SeedSelector = shoot.Spec.SeedSelector
}
