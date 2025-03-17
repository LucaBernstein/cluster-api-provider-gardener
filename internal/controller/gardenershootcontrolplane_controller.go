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

package controller

import (
	"context"
	"time"

	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/v1alpha1"
)

// GardenerShootControlPlaneReconciler reconciles a GardenerShootControlPlane object
type GardenerShootControlPlaneReconciler struct {
	Client         client.Client
	GardenerClient client.Client
	Scheme         *runtime.Scheme
	Log            logr.Logger

	shootControlPlane *infrav1alpha1.GardenerShootControlPlane
	shoot             *gardenercorev1beta1.Shoot
}

// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GardenerShootControlPlane object against the actual control plane state, and then
// perform operations to make the control plane state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *GardenerShootControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = runtimelog.FromContext(ctx).WithValues("gardenershootcontrolplane", req.NamespacedName)

	r.Log.Info("Getting GardenerShootControlPlane object")
	r.shootControlPlane = &infrav1alpha1.GardenerShootControlPlane{}
	if err := r.Client.Get(ctx, req.NamespacedName, r.shootControlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("resource no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	r.Log.Info("Getting own cluster")
	cluster, err := util.GetOwnerCluster(ctx, r.Client, r.shootControlPlane.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		r.Log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{Requeue: true}, nil
	}

	shootNamespace := r.shootControlPlane.Namespace
	if len(r.shootControlPlane.Project) > 0 {
		shootNamespace = "garden-" + r.shootControlPlane.Project
	}
	r.shoot = &gardenercorev1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.shootControlPlane.Name,
			Namespace: shootNamespace,
		},
		Spec: r.shootControlPlane.Spec,
	}

	// Handle deleted clusters
	if !r.shootControlPlane.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx)
	}

	// Handle non-deleted clusters
	return r.reconcile(ctx)
}

func (r *GardenerShootControlPlaneReconciler) reconcile(ctx context.Context) (ctrl.Result, error) {
	r.Log.Info("Reconciling GardenerShootControlPlane")

	r.Log.Info("Adding finalizer to GardenerShootControlPlane")
	patch := client.MergeFrom(r.shootControlPlane.DeepCopy())
	if controllerutil.AddFinalizer(r.shootControlPlane, v1beta1.ClusterFinalizer) {
		if err := r.Client.Patch(ctx, r.shootControlPlane, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	err := r.GardenerClient.Get(ctx, client.ObjectKeyFromObject(r.shoot), r.shoot)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("Shoot not found, creating it")
			if err := r.GardenerClient.Create(ctx, r.shoot); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	if isReady, err := r.patchStatus(ctx, r.shoot); err != nil {
		return ctrl.Result{Requeue: !isReady, RequeueAfter: 30 * time.Second}, err
	}

	r.Log.Info("Successfully reconciled GardenerShootControlPlane")
	record.Event(r.shootControlPlane, "GardenerShootControlPlaneReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

func (r *GardenerShootControlPlaneReconciler) reconcileDelete(ctx context.Context) (ctrl.Result, error) {
	r.Log.Info("Reconciling Delete GardenerShootControlPlane")

	err := r.GardenerClient.Get(ctx, client.ObjectKeyFromObject(r.shoot), r.shoot)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		r.Log.Info("Shoot not found")
	}

	if _, err := r.patchStatus(ctx, r.shoot); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if !apierrors.IsNotFound(err) {
		patch := client.MergeFrom(r.shoot.DeepCopy())
		annotations.AddAnnotations(r.shoot, map[string]string{constants.ConfirmationDeletion: "true"})
		if err := r.GardenerClient.Patch(ctx, r.shoot, patch); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err := r.GardenerClient.Delete(ctx, r.shoot); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Minute}, nil
	}

	patch := client.MergeFrom(r.shootControlPlane.DeepCopy())
	if controllerutil.RemoveFinalizer(r.shootControlPlane, v1beta1.ClusterFinalizer) {
		err := r.Client.Patch(ctx, r.shootControlPlane, patch)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	r.Log.Info("Successfully reconciled deletion of GardenerShootControlPlane")
	record.Event(r.shootControlPlane, "GardenerShootControlPlaneReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

func (r *GardenerShootControlPlaneReconciler) patchStatus(ctx context.Context, shoot *gardenercorev1beta1.Shoot) (bool, error) {
	patch := client.MergeFrom(r.shootControlPlane.DeepCopy())
	if shoot != nil {
		shootStatus := gardener.ComputeShootStatus(shoot.Status.LastOperation, shoot.Status.LastErrors, shoot.Status.Conditions...)
		r.shootControlPlane.Status.Ready = shootStatus == gardener.ShootStatusHealthy
	}
	return r.shootControlPlane.Status.Ready, r.Client.Status().Patch(ctx, r.shootControlPlane, patch)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GardenerShootControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.GardenerShootControlPlane{}).
		Named("gardenershootcontrolplane").
		Complete(r)
}
