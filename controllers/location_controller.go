/*
Copyright 2023 mgrote.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/internal"
)

// LocationReconciler reconciles a Location object
type LocationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=locations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=locations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=locations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Location object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *LocationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("LocationReconciler", req.NamespacedName)

	location := &personaliotv1alpha1.Location{}
	if err := r.Get(ctx, req.NamespacedName, location); err != nil {
		logger.Error(err, "unable to fetch power outlet")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	powerStripList := &personaliotv1alpha1.PowerstripList{}
	stripListOpts := &client.ListOptions{
		Namespace:     location.Namespace,
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.location": location.Name}),
	}

	var outletList []*personaliotv1alpha1.Poweroutlet
	for _, strip := range powerStripList.Items {
		for _, outlet := range strip.Spec.Outlets {
			outletList = append(outletList, outlet)
		}
	}

	// Check if deletion timestamp is added, switch off power outlet and delete finalizer.
	if !location.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, outletList, location)
	}

	// Check if finalizer should be added or applied.
	if !controllerutil.ContainsFinalizer(location, personaliotv1alpha1.LocationFinalizer) {
		controllerutil.AddFinalizer(location, personaliotv1alpha1.LocationFinalizer)
		if err := r.Update(ctx, location); err != nil {
			return ctrl.Result{}, err
		}
	}

	// No mood is set, nothing to do.
	if location.Spec.Mood == "" {
		return ctrl.Result{}, nil
	}

	if err := r.List(ctx, powerStripList, stripListOpts); err != nil {
		return ctrl.Result{}, err
	}

	if len(outletList) == 0 {
		return ctrl.Result{}, nil
	}

	switch location.Spec.Mood {
	case personaliotv1alpha1.LocationMoodDark:
		return r.reconcileDark(ctx, outletList, location)
	case personaliotv1alpha1.LocationMoodBright:
		return r.reconcileBright(ctx, outletList, location)
	case personaliotv1alpha1.LocationMoodDontKnow:
		return r.reconcileDontKnow(ctx, outletList, location)
	}

	return ctrl.Result{}, fmt.Errorf("location mood %s not recognised", location.Spec.Mood)
}

func (r *LocationReconciler) reconcileDark(ctx context.Context, outlets []*personaliotv1alpha1.Poweroutlet, location *personaliotv1alpha1.Location) (ctrl.Result, error) {
	for _, outlet := range outlets {
		if outlet.Spec.Switch == internal.PowerOnSignal {
			outlet.Spec.Switch = internal.PowerOffSignal
			if err := r.Update(ctx, outlet); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	location.Status.Mood = personaliotv1alpha1.LocationMoodDark
	err := r.Update(ctx, location)
	return ctrl.Result{}, err
}

func (r *LocationReconciler) reconcileBright(ctx context.Context, outlets []*personaliotv1alpha1.Poweroutlet, location *personaliotv1alpha1.Location) (ctrl.Result, error) {
	for _, outlet := range outlets {
		if outlet.Spec.Switch == internal.PowerOffSignal {
			outlet.Spec.Switch = internal.PowerOnSignal
			if err := r.Update(ctx, outlet); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	location.Status.Mood = personaliotv1alpha1.LocationMoodBright
	err := r.Update(ctx, location)
	return ctrl.Result{}, err
}

func (r *LocationReconciler) reconcileDontKnow(ctx context.Context, outlets []*personaliotv1alpha1.Poweroutlet, location *personaliotv1alpha1.Location) (ctrl.Result, error) {
	for _, outlet := range outlets {
		if outlet.Spec.Switch == internal.PowerOffSignal {
			outlet.Spec.Switch = internal.PowerOnSignal
			if err := r.Update(ctx, outlet); err != nil {
				return ctrl.Result{}, err
			}
		}
		if outlet.Spec.Switch == internal.PowerOnSignal {
			outlet.Spec.Switch = internal.PowerOffSignal
			if err := r.Update(ctx, outlet); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	location.Status.Mood = personaliotv1alpha1.LocationMoodDontKnow
	err := r.Update(ctx, location)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&personaliotv1alpha1.Location{}).
		Complete(r)
}

func (r *LocationReconciler) reconcileDelete(ctx context.Context, outlets []*personaliotv1alpha1.Poweroutlet, location *personaliotv1alpha1.Location) (ctrl.Result, error) {
	// turn the lights off before you go
	_, err := r.reconcileDark(ctx, outlets, location)
	if err != nil {
		return ctrl.Result{}, err
	}
	controllerutil.RemoveFinalizer(location, personaliotv1alpha1.LocationFinalizer)
	if err = r.Update(ctx, location); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
