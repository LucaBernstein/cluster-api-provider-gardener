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

	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/v1alpha1"
)

// GardenerShootClusterReconciler reconciles a GardenerShootCluster object
type GardenerShootClusterReconciler struct {
	Client         client.Client
	GardenerClient client.Client
	Scheme         *runtime.Scheme
	Log            logr.Logger
}

// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gardenershootclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gardenershootclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=gardenershootclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GardenerShootCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *GardenerShootClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx).WithValues("gardenershootcluster", req.NamespacedName)

	var (
		shootCluster infrav1alpha1.GardenerShootCluster
		shoot        gardenercorev1beta1.Shoot
	)
	if err := r.Client.Get(ctx, req.NamespacedName, &shootCluster); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("resource no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cluster, err := util.GetOwnerCluster(ctx, r.Client, shootCluster.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cluster == nil {
		r.Log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	err = r.GardenerClient.Get(ctx, types.NamespacedName{
		Name:      shootCluster.Name,
		Namespace: shootCluster.Namespace,
	}, &shoot)

	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("Wo Shoot?")
			shoot := gardenercorev1beta1.Shoot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      shootCluster.Name,
					Namespace: shootCluster.Namespace,
				},
				Spec: shootCluster.Spec,
			}
			err = r.GardenerClient.Create(ctx, &shoot)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GardenerShootClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.GardenerShootCluster{}).
		Named("gardenershootcluster").
		Complete(r)
}
