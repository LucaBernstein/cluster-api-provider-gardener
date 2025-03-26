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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	controlplanev1alpha1 "github.com/gardener/cluster-api-provider-gardener/api/v1alpha1"
	"github.com/gardener/cluster-api-provider-gardener/internal/controller"
)

// nolint:unused
// log is for logging in this package.
var _ = logf.Log.WithName("gardenershootcontrolplane-resource")

// SetupGardenerShootControlPlaneWebhookWithManager registers the webhook for GardenerShootControlPlane in the manager.
func SetupGardenerShootControlPlaneWebhookWithManager(mgr ctrl.Manager, gardenerClient client.Client) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&controlplanev1alpha1.GardenerShootControlPlane{}).
		WithValidator(&GardenerShootControlPlaneCustomValidator{
			GardenerClient: gardenerClient,
		}).
		WithDefaulter(&GardenerShootControlPlaneCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-controlplane-cluster-x-k8s-io-v1alpha1-gardenershootcontrolplane,mutating=true,failurePolicy=fail,sideEffects=None,groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes,verbs=create;update,versions=v1alpha1,name=vgardenershootcontrolplane-v1alpha1.kb.io,admissionReviewVersions=v1

// GardenerShootControlPlaneCustomDefaulter struct is responsible for defaulting the GardenerShootControlPlane resource.
type GardenerShootControlPlaneCustomDefaulter struct {
	GardenerClient client.Client
}

var _ webhook.CustomDefaulter = &GardenerShootControlPlaneCustomDefaulter{}

func (d GardenerShootControlPlaneCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	shootControlPlane, ok := obj.(*controlplanev1alpha1.GardenerShootControlPlane)
	if !ok {
		return fmt.Errorf("expected a GardenerShootControlPlane object for the obj but got %T", obj)
	}

	if len(shootControlPlane.Spec.ProjectNamespace) == 0 {
		shootControlPlane.Spec.ProjectNamespace = shootControlPlane.Namespace
	}

	return nil
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-controlplane-cluster-x-k8s-io-v1alpha1-gardenershootcontrolplane,mutating=false,failurePolicy=fail,sideEffects=None,groups=controlplane.cluster.x-k8s.io,resources=gardenershootcontrolplanes,verbs=create;update,versions=v1alpha1,name=vgardenershootcontrolplane-v1alpha1.kb.io,admissionReviewVersions=v1

// GardenerShootControlPlaneCustomValidator struct is responsible for validating the GardenerShootControlPlane resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type GardenerShootControlPlaneCustomValidator struct {
	GardenerClient client.Client
}

var _ webhook.CustomValidator = &GardenerShootControlPlaneCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type GardenerShootControlPlane.
func (v *GardenerShootControlPlaneCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	shootControlPlane, ok := obj.(*controlplanev1alpha1.GardenerShootControlPlane)
	if !ok {
		return nil, fmt.Errorf("expected a GardenerShootControlPlane object but got %T", obj)
	}

	return nil, v.GardenerClient.Create(ctx, controller.ShootFromControlPlane(shootControlPlane), &client.CreateOptions{DryRun: []string{"All"}})
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type GardenerShootControlPlane.
func (v *GardenerShootControlPlaneCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	shootControlPlane, ok := newObj.(*controlplanev1alpha1.GardenerShootControlPlane)
	if !ok {
		return nil, fmt.Errorf("expected a GardenerShootControlPlane object for the newObj but got %T", newObj)
	}

	// For the update, we need to get the actual cluster and inject the new config, because e.g. the resourceVersion must be set.
	shoot := controller.ShootFromControlPlane(shootControlPlane)
	if err := v.GardenerClient.Get(ctx, client.ObjectKeyFromObject(shoot), shoot); err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	shoot.Spec = shootControlPlane.Spec.ShootSpec

	// During deletion, it can happen that the Shoot wants to be patched, when it does not exist anymore,
	// therefore ignoring this error to prevent the reconciliation to be blocked.
	return nil, client.IgnoreNotFound(v.GardenerClient.Update(ctx, shoot, &client.UpdateOptions{DryRun: []string{"All"}}))
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type GardenerShootControlPlane.
func (v *GardenerShootControlPlaneCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*controlplanev1alpha1.GardenerShootControlPlane)
	if !ok {
		return nil, fmt.Errorf("expected a GardenerShootControlPlane object but got %T", obj)
	}

	return nil, nil
}
