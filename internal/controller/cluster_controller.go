package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/kcp"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gardener/cluster-api-provider-gardener/api/v1alpha1"
)

// ClusterController mocks the cluster-api Cluster controller.
// This _ONLY_ works with the Gardener provider, as no dynamic watching is being done here.
type ClusterController struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *ClusterController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := runtimelog.FromContext(ctx).WithValues("cluster-object", req.NamespacedName, "cluster", req.ClusterName)

	log.Info("Getting Cluster")
	cluster := v1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("resource no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Mocking setting the Owner reference for GardenerShootControlPlanes
	gscp := &v1alpha1.GardenerShootControlPlane{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      cluster.Spec.ControlPlaneRef.Name,
		Namespace: cluster.Spec.ControlPlaneRef.Namespace,
	}, gscp); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("could not find respective GSCP. Requeueing.")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	if err := ensureOwnerRef(ctx, r.Client, gscp, &cluster); err != nil {
		log.Error(err, "unable to ensure OwnerRef on GSCP")
		return ctrl.Result{}, err
	}

	cluster.Status = v1beta1.ClusterStatus{
		Phase:               string(v1beta1.ClusterPhaseProvisioned),
		InfrastructureReady: true,
		ControlPlaneReady:   gscp.Status.Initialized,
		ObservedGeneration:  cluster.Generation,
	}
	if !gscp.Status.Initialized {
		cluster.Status.Phase = string(v1beta1.ClusterPhaseProvisioning)
	}
	if err := r.Client.Status().Update(ctx, &cluster); err != nil {
		log.Error(err, "unable to update cluster status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func ensureOwnerRef(ctx context.Context, c client.Client, obj *v1alpha1.GardenerShootControlPlane, cluster *v1beta1.Cluster) error {
	desiredOwnerRef := metav1.OwnerReference{
		APIVersion: cluster.APIVersion,
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if util.HasExactOwnerRef(obj.GetOwnerReferences(), desiredOwnerRef) &&
		obj.GetLabels()[v1beta1.ClusterNameLabel] == cluster.Name {
		return nil
	}

	if err := controllerutil.SetControllerReference(cluster, obj, c.Scheme()); err != nil {
		return err
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[v1beta1.ClusterNameLabel] = cluster.Name
	obj.SetLabels(labels)

	return c.Update(ctx, obj)
}

func (r *ClusterController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Cluster{}).
		Named("cluster").
		Owns(&v1alpha1.GardenerShootControlPlane{}).
		Complete(kcp.WithClusterInContext(r))
}
