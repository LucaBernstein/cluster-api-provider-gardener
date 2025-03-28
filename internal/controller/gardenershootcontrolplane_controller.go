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
	"fmt"
	"time"

	gardenerauthenticationv1alpha1 "github.com/gardener/gardener/pkg/apis/authentication/v1alpha1"
	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	controlplanev1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/v1alpha1"
)

// GardenerShootControlPlaneReconciler reconciles a GardenerShootControlPlane object
type GardenerShootControlPlaneReconciler struct {
	Client         client.Client
	GardenerClient client.Client
	Scheme         *runtime.Scheme
}

type ControlPlaneContext struct {
	log logr.Logger
	ctx context.Context

	cluster           *v1beta1.Cluster
	shootControlPlane *controlplanev1alpha1.GardenerShootControlPlane
	shoot             *gardenercorev1beta1.Shoot
}

// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.gardener.cloud,resources=shoots/adminkubeconfig,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=core.gardener.cloud,resources=shoots;shoots/status,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *GardenerShootControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := runtimelog.FromContext(ctx).WithValues("gardenershootcontrolplane", req.NamespacedName)

	cpc := ControlPlaneContext{
		log: log,
		ctx: ctx,
	}

	log.Info("Getting GardenerShootControlPlane object")
	cpc.shootControlPlane = &controlplanev1alpha1.GardenerShootControlPlane{}
	if err := r.Client.Get(cpc.ctx, req.NamespacedName, cpc.shootControlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("resource no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Getting own cluster")
	var err error
	cpc.cluster, err = util.GetOwnerCluster(cpc.ctx, r.Client, cpc.shootControlPlane.ObjectMeta)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cpc.cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return ctrl.Result{Requeue: true}, nil
	}

	cpc.shoot = ShootFromControlPlane(cpc.shootControlPlane)

	// Handle deleted clusters
	if !cpc.shootControlPlane.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(cpc)
	}

	// Handle non-deleted clusters
	return r.reconcile(cpc)
}

func (r *GardenerShootControlPlaneReconciler) reconcile(cpc ControlPlaneContext) (ctrl.Result, error) {
	log := cpc.log
	log.Info("Reconciling GardenerShootControlPlane")

	log.Info("Adding finalizer to GardenerShootControlPlane")
	patch := client.MergeFrom(cpc.shootControlPlane.DeepCopy())
	if controllerutil.AddFinalizer(cpc.shootControlPlane, v1beta1.ClusterFinalizer) {
		if err := r.Client.Patch(cpc.ctx, cpc.shootControlPlane, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	err := r.GardenerClient.Get(cpc.ctx, client.ObjectKeyFromObject(cpc.shoot), cpc.shoot)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		log.Info("Shoot not found, creating it")
		if err := r.GardenerClient.Create(cpc.ctx, cpc.shoot); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.updateStatus(cpc); err != nil {
		log.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	if err := r.syncControlPlaneSpecs(cpc); err != nil {
		log.Error(err, "failed to sync control plane spec")
		return ctrl.Result{}, err
	}

	if !cpc.shootControlPlane.Status.Initialized {
		// Wait until the shoot is initialized.
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	log.Info("Reconcile Shoot Access for ClusterAPI")
	err = r.reconcileShootAccess(cpc)
	if err != nil {
		log.Error(err, "Error reconciling Shoot Access for ClusterAPI")
		return ctrl.Result{}, err
	}

	log.Info("Reconcile shootControlEndpoint")
	err = r.reconcileShootControlPlaneEndpoint(cpc)
	if err != nil {
		log.Error(err, "Error reconciling shootControlEndpoint")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled GardenerShootControlPlane")
	record.Event(cpc.shootControlPlane, "GardenerShootControlPlaneReconcile", "Reconciled")
	if !cpc.shootControlPlane.Status.Ready {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

func (r *GardenerShootControlPlaneReconciler) reconcileDelete(cpc ControlPlaneContext) (ctrl.Result, error) {
	log := cpc.log
	log.Info("Reconciling Delete GardenerShootControlPlane")

	err := r.Client.Delete(cpc.ctx, newEmptyShootAccessSecret(cpc.cluster))
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		log.Info("Shoot Access Secret not found")
	}

	err = r.GardenerClient.Get(cpc.ctx, client.ObjectKeyFromObject(cpc.shoot), cpc.shoot)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		log.Info("Shoot not found")
		cpc.shoot = nil
	}

	if err = r.updateStatus(cpc); err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if cpc.shoot != nil {
		// Propagate the deletion to the shoot.
		patch := client.MergeFrom(cpc.shoot.DeepCopy())
		annotations.AddAnnotations(cpc.shoot, map[string]string{constants.ConfirmationDeletion: "true"})
		if err := r.GardenerClient.Patch(cpc.ctx, cpc.shoot, patch); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		if err := r.GardenerClient.Delete(cpc.ctx, cpc.shoot); err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Minute}, nil
	}

	patch := client.MergeFrom(cpc.shootControlPlane.DeepCopy())
	if controllerutil.RemoveFinalizer(cpc.shootControlPlane, v1beta1.ClusterFinalizer) {
		if err = r.Client.Patch(cpc.ctx, cpc.shootControlPlane, patch); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Successfully reconciled deletion of GardenerShootControlPlane")
	record.Event(cpc.shootControlPlane, "GardenerShootControlPlaneReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

func (r *GardenerShootControlPlaneReconciler) reconcileShootControlPlaneEndpoint(cpc ControlPlaneContext) error {
	endpoint := ""
	for _, address := range cpc.shoot.Status.AdvertisedAddresses {
		if address.Name == constants.AdvertisedAddressExternal {
			endpoint = address.URL
			break
		}
	}

	if len(endpoint) == 0 {
		return fmt.Errorf("could not find external advertised address for shoot")
	}

	patch := client.MergeFrom(cpc.shootControlPlane.DeepCopy())
	cpc.shootControlPlane.Spec.ControlPlaneEndpoint = clusterv1beta1.APIEndpoint{
		Host: endpoint,
		Port: 443,
	}
	return r.Client.Patch(cpc.ctx, cpc.shootControlPlane, patch)
}

func (r *GardenerShootControlPlaneReconciler) reconcileShootAccess(cpc ControlPlaneContext) error {
	secret := newEmptyShootAccessSecret(cpc.cluster)
	err := r.Client.Get(cpc.ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		// Create (empty secret)
		if err = r.Client.Create(cpc.ctx, secret); err != nil {
			return err
		}
	}

	adminKubeconfigRequest := &gardenerauthenticationv1alpha1.AdminKubeconfigRequest{
		Spec: gardenerauthenticationv1alpha1.AdminKubeconfigRequestSpec{
			ExpirationSeconds: ptr.To(int64(6000)),
		},
	}
	if err := r.Client.SubResource("adminkubeconfig").Create(cpc.ctx, cpc.shoot, adminKubeconfigRequest); err != nil {
		return err
	}

	secret.Data = map[string][]byte{"value": adminKubeconfigRequest.Status.Kubeconfig}

	return r.Client.Update(cpc.ctx, secret)
}

func newEmptyShootAccessSecret(cluster *v1beta1.Cluster) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kubeconfig", cluster.Name),
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name": cluster.Name,
			},
		},
		Type: v1beta1.ClusterSecretType,
	}
}

func (r *GardenerShootControlPlaneReconciler) syncControlPlaneSpecs(cpc ControlPlaneContext) error {
	var (
		err error

		originalShoot          = cpc.shoot.DeepCopy()
		patchShoot             = client.MergeFrom(originalShoot)
		patchShootControlPlane = client.MergeFrom(cpc.shootControlPlane.DeepCopy())
	)

	// patch the shoot cluster object from the GardenerShootControlPlane object.
	cpc.log.Info("Syncing GardenerShootControlPlane spec >>> Shoot spec")
	cpc.shoot.Spec = cpc.shootControlPlane.Spec.ShootSpec

	if kubernetesVersion := cpc.shootControlPlane.Spec.Version; len(kubernetesVersion) > 0 {
		cpc.shoot.Spec.Kubernetes.Version = kubernetesVersion
	}

	err = r.GardenerClient.Patch(cpc.ctx, cpc.shoot, patchShoot)
	cpc.shootControlPlane.Spec.ShootSpec = cpc.shoot.Spec
	if err != nil {
		cpc.log.Error(err, "Error while syncing GardenerShootControlPlane to Gardener Cluster Shoot")
		cpc.shootControlPlane.Spec.ShootSpec = originalShoot.Spec
	}

	// sync back the shoot state (also, if above sync failed).
	cpc.log.Info("Syncing GardenerShootControlPlane spec <<< Shoot spec")
	err1 := r.Client.Patch(cpc.ctx, cpc.shootControlPlane, patchShootControlPlane)
	if err1 != nil {
		cpc.log.Error(err, "Error while syncing Gardener Cluster Shoot to GardenerShootControlPlane")
		if err != nil {
			return err
		}
	}
	return err1
}

func (r *GardenerShootControlPlaneReconciler) updateStatus(cpc ControlPlaneContext) error {
	if cpc.shoot != nil {
		shootStatus := gardener.ComputeShootStatus(cpc.shoot.Status.LastOperation, cpc.shoot.Status.LastErrors, cpc.shoot.Status.Conditions...)
		// TODO(LucaBernstein): Adapt readiness check to assert shoot component conditions.
		cpc.shootControlPlane.Status.Ready = shootStatus == gardener.ShootStatusHealthy
		if !cpc.shootControlPlane.Status.Initialized {
			cpc.shootControlPlane.Status.Initialized = controlPlaneReady(cpc.shoot.Status)
		}
		cpc.shootControlPlane.Status.ShootStatus = cpc.shoot.Status
	}
	return r.Client.Status().Update(cpc.ctx, cpc.shootControlPlane)
}

func controlPlaneReady(shootStatus gardenercorev1beta1.ShootStatus) bool {
	for _, condition := range shootStatus.Conditions {
		if condition.Type != gardenercorev1beta1.ShootControlPlaneHealthy {
			continue
		}
		if condition.Status == gardenercorev1beta1.ConditionTrue {
			return true
		}
	}
	return false
}

func ShootFromControlPlane(shootControlPlane *controlplanev1alpha1.GardenerShootControlPlane) *gardenercorev1beta1.Shoot {
	return &gardenercorev1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shootControlPlane.Name,
			Namespace: shootControlPlane.Spec.ProjectNamespace,
		},
		Spec: shootControlPlane.Spec.ShootSpec,
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *GardenerShootControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager, targetCluster cluster.Cluster) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controlplanev1alpha1.GardenerShootControlPlane{}).
		Named("gardenershootcontrolplane").
		WatchesRawSource(
			source.Kind[client.Object](targetCluster.GetCache(),
				&gardenercorev1beta1.Shoot{},
				handler.EnqueueRequestsFromMapFunc(r.MapShootToControlPlaneObject),
			),
		).
		Complete(r)
}

func (r *GardenerShootControlPlaneReconciler) MapShootToControlPlaneObject(ctx context.Context, obj client.Object) []reconcile.Request {
	var (
		log = runtimelog.FromContext(ctx).WithValues("shoot", client.ObjectKeyFromObject(obj))

		controlPlaneList controlplanev1alpha1.GardenerShootControlPlaneList
	)
	if err := r.Client.List(ctx, &controlPlaneList, client.MatchingFields{controlplanev1alpha1.ShootReferenceIndexKey: client.ObjectKeyFromObject(obj).String()}); err != nil || len(controlPlaneList.Items) != 1 {
		log.Error(err, "Could not list control planes")
		return nil
	}
	return []reconcile.Request{{NamespacedName: client.ObjectKeyFromObject(&controlPlaneList.Items[0])}}
}
