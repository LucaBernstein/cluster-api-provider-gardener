package infrastructure

import (
	"context"
	"fmt"
	controlplanev1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/controlplane/v1alpha1"
	providerutil "github.com/gardener/cluster-api-provider-gardener/internal/util"
	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/infrastructure/v1alpha1"
)

type GardenerShootClusterReconciler struct {
	Client         client.Client
	GardenerClient client.Client
	Scheme         *runtime.Scheme
	IsKCP          bool
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gardenershootclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gardenershootclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gardenershootclusters/finalizers,verbs=update

func (r *GardenerShootClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := runtimelog.FromContext(ctx).WithValues("gardenershootcluster", req.NamespacedName, "cluster", req.ClusterName)

	log.Info("Reconciling GardenerShootCluster")
	infraCluster := &infrastructurev1alpha1.GardenerShootCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, infraCluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("GardenerShootCluster not found or already deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get GardenerShootCluster")
		return ctrl.Result{}, err
	}

	cluster, err := util.GetOwnerCluster(ctx, r.Client, infraCluster.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to get owner Cluster")
		return ctrl.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, infraCluster) {
		log.Info("GardenerShootCluster or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	if !infraCluster.DeletionTimestamp.IsZero() {
		log.Info("GardenerShootCluster is being deleted")
		return r.reconcileDelete(ctx, infraCluster)
	}

	return r.reconcile(ctx, infraCluster, cluster)
}

func (r *GardenerShootClusterReconciler) reconcileDelete(ctx context.Context, infraCluster *infrastructurev1alpha1.GardenerShootCluster) (ctrl.Result, error) {
	log := runtimelog.FromContext(ctx).WithValues("gardenershootcluster", client.ObjectKeyFromObject(infraCluster), "operation", "delete")

	patch := client.MergeFrom(infraCluster.DeepCopy())
	if controllerutil.RemoveFinalizer(infraCluster, v1beta1.ClusterFinalizer) {
		if err := r.Client.Patch(ctx, infraCluster, patch); err != nil {
			log.Error(err, "Failed to patch GardenerShootCluster finalizer")
			return ctrl.Result{}, err
		}
	}

	log.Info("GardenerShootCluster deleted successfully")
	return ctrl.Result{}, nil
}

func (r *GardenerShootClusterReconciler) reconcile(ctx context.Context, infraCluster *infrastructurev1alpha1.GardenerShootCluster, cluster *v1beta1.Cluster) (ctrl.Result, error) {
	log := runtimelog.FromContext(ctx).WithValues("gardenershootcluster", client.ObjectKeyFromObject(infraCluster), "operation", "reconcile")

	patch := client.MergeFrom(infraCluster.DeepCopy())
	if controllerutil.AddFinalizer(infraCluster, v1beta1.ClusterFinalizer) {
		if err := r.Client.Patch(ctx, infraCluster, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.syncSpecs(ctx, infraCluster, cluster); err != nil {
		log.Error(err, "Failed to sync GardenerShootCluster spec")
		return ctrl.Result{}, err
	}

	if !infraCluster.Status.Ready {
		infraCluster.Status.Ready = true
		if err := r.Client.Status().Patch(ctx, infraCluster, patch); err != nil {
			log.Error(err, "Failed to patch GardenerShootCluster status")
			return ctrl.Result{}, err
		}
	}
	log.Info("GardenerShootCluster reconciled successfully")
	return ctrl.Result{}, nil
}

func (r *GardenerShootClusterReconciler) syncSpecs(ctx context.Context, infraCluster *infrastructurev1alpha1.GardenerShootCluster, cluster *v1beta1.Cluster) error {
	log := runtimelog.FromContext(ctx).WithValues("gardenershootcluster", client.ObjectKeyFromObject(infraCluster), "operation", "syncSpecs")

	if cluster.Spec.ControlPlaneRef == nil {
		log.Info("ControlPlaneRef is nil, do nothing")
		return nil
	}
	controlPlane := &controlplanev1alpha1.GardenerShootControlPlane{}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: cluster.Spec.ControlPlaneRef.Namespace, Name: cluster.Spec.ControlPlaneRef.Name}, controlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ControlPlane not found")
			return nil
		}
		log.Error(err, "Failed to get ControlPlane")
		return err
	}

	shoot := &gardenercorev1beta1.Shoot{}
	if err := r.GardenerClient.Get(ctx, providerutil.ShootNameFromCAPIResources(*cluster, *controlPlane), shoot); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	var (
		originalShoot = shoot.DeepCopy()
		patchShoot    = client.StrategicMergeFrom(originalShoot)

		originalInfraCluster = infraCluster.DeepCopy()
		patchInfraCluster    = client.MergeFrom(originalInfraCluster)
	)

	// Cross-patch Shoot and GardenerShootCluster objects.
	providerutil.SyncShootSpecFromCluster(shoot, originalInfraCluster)
	providerutil.SyncClusterSpecFromShoot(originalShoot, infraCluster)

	log.Info("Syncing GardenerShootCluster spec >>> Shoot spec")
	if err := r.GardenerClient.Patch(ctx, shoot, patchShoot); err != nil {
		log.Error(err, "Error while syncing GardenerShootCluster to Gardener Shoot")
	}

	// sync back the shoot state (also, if above sync failed)
	log.Info("Syncing GardenerShootCluster spec <<< Shoot spec")
	if err := r.Client.Patch(ctx, infraCluster, patchInfraCluster); err != nil {
		log.Error(err, "Error while syncing Gardener Shoot to GardenerShootCluster")
		return err
	}

	return nil
}

func (r *GardenerShootClusterReconciler) SetupWithManager(mgr ctrl.Manager, targetCluster cluster.Cluster) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha1.GardenerShootCluster{}).
		Named("gardenershootcluster").
		WatchesRawSource(
			source.Kind[client.Object](targetCluster.GetCache(),
				&gardenercorev1beta1.Shoot{},
				handler.EnqueueRequestsFromMapFunc(r.MapShootToGardenerShootClusterObject),
			),
		).
		Complete(r)
}

func (r *GardenerShootClusterReconciler) MapShootToGardenerShootClusterObject(ctx context.Context, obj client.Object) []reconcile.Request {
	var (
		log          = runtimelog.FromContext(ctx).WithValues("shoot", client.ObjectKeyFromObject(obj))
		clusterName  string
		infraCluster *infrastructurev1alpha1.GardenerShootCluster
	)
	shoot, ok := obj.(*gardenercorev1beta1.Shoot)
	if !ok {
		log.Error(fmt.Errorf("could not assert object to Shoot"), "")
		return nil
	}

	namespace, ok := shoot.GetLabels()[infrastructurev1alpha1.GSCReferenceNamespaceKey]
	if !ok {
		log.Info("Could not find gsc namespace on label")
		return nil
	}

	name, ok := shoot.GetLabels()[infrastructurev1alpha1.GSCReferenceNameKey]
	if !ok {
		log.Info("Could not find gsc name on label")
	}
	if r.IsKCP {
		clusterName, ok = shoot.GetLabels()[infrastructurev1alpha1.GSCReferecenceClusterNameKey]
		if !ok {
			log.Info("Could not find gsc cluster on label")
		}
	}

	infraCluster = &infrastructurev1alpha1.GardenerShootCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(infraCluster), ClusterName: clusterName}}
}
