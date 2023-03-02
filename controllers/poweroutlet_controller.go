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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/internal"
	"github.com/mgrote/personal-iot/internal/mqttiot"
)

// PoweroutletReconciler reconciles a Poweroutlet object
type PoweroutletReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	MQTTSubscriber mqttiot.MQTTSubscriber
	MQTTPublisher  mqttiot.MQTTPublisher
}

//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=poweroutlets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=poweroutlets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=poweroutlets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Poweroutlet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *PoweroutletReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("PoweroutletReconciler", req.NamespacedName)

	powerOutlet := &personaliotv1alpha1.Poweroutlet{}
	if err := r.Get(ctx, req.NamespacedName, powerOutlet); err != nil {
		logger.Error(err, "unable to fetch power outlet")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.WithValues("switch", powerOutlet.Spec.Switch).Info("found power outlet")

	// Check if deletion timestamp is added, switch off power outlet and delete finalizer.
	if !powerOutlet.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, powerOutlet)
	}

	// Check if finalizer should be added or applied.
	if !controllerutil.ContainsFinalizer(powerOutlet, personaliotv1alpha1.PowerOutletFinalizer) {
		controllerutil.AddFinalizer(powerOutlet, personaliotv1alpha1.PowerOutletFinalizer)
		if err := r.Update(ctx, powerOutlet); err != nil {
			return ctrl.Result{}, err
		}
	}

	// if nothing is to do, leave
	if powerOutlet.Spec.Switch == powerOutlet.Status.Switch {
		logger.Info("desired switch state reached, nothing else to do", "", powerOutlet.Spec.Switch, "state", powerOutlet.Spec.Switch)
		return ctrl.Result{}, nil
	}

	currentState, err := r.reconcilePowerOutletState(ctx, powerOutlet)
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("reached state", "current state", currentState)
	powerOutlet.Status.Switch = *currentState

	if err = r.Status().Update(ctx, powerOutlet); err != nil {
		logger.Error(err, "update PowerOutlet status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PoweroutletReconciler) reconcileDelete(ctx context.Context, powerOutlet *personaliotv1alpha1.Poweroutlet) (ctrl.Result, error) {
	if powerOutlet.Status.Switch == internal.PowerOnSignal {
		powerOutlet.Spec.Switch = internal.PowerOffSignal

		currentState, err := r.reconcilePowerOutletState(ctx, powerOutlet)
		if err != nil {
			return ctrl.Result{}, err
		}
		if *currentState != internal.PowerOffSignal {
			return ctrl.Result{}, fmt.Errorf("finalize can not switch off device %s", powerOutlet.Name)
		}

		powerOutlet.Status.Switch = *currentState
		if err = r.Status().Update(ctx, powerOutlet); err != nil {
			return ctrl.Result{}, err
		}
	}
	controllerutil.RemoveFinalizer(powerOutlet, personaliotv1alpha1.PowerOutletFinalizer)
	if err := r.Update(ctx, powerOutlet); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *PoweroutletReconciler) reconcilePowerOutletState(ctx context.Context, powerOutlet *personaliotv1alpha1.Poweroutlet) (*string, error) {
	if err := r.MQTTPublisher.Connect(); err != nil {
		return nil, err
	}

	// request desired state from outlet
	if err := r.MQTTPublisher.Publish(powerOutlet.Spec.MQTTCommandTopik, powerOutlet.Spec.Switch, 1, true); err != nil {
		return nil, err
	}
	r.MQTTPublisher.Disconnect(500)

	// check if state was reached, keep in mind: there are some timing problems
	if err := r.MQTTSubscriber.Connect(); err != nil {
		return nil, err
	}

	messageChannel := make(chan mqttiot.MQTTMessage)
	if err := r.MQTTSubscriber.Subscribe(powerOutlet.Spec.MQTTStatusTopik, 1, messageChannel); err != nil {
		return nil, err
	}

	// Wait for switch state change.
	// When subscribing the power outlet status topik, the first message delivers immediately the current state.
	// The status change may come later, so we have to wait for the next messages.
	var currentState string
	for i := 0; i < 10; i++ {
		incoming := <-messageChannel
		currentState = incoming.Msg
		// TODO output improves the timing, think about a wait ;-).
		fmt.Println("current state", currentState)
		if currentState == powerOutlet.Spec.Switch {
			break
		}
	}

	r.MQTTSubscriber.Disconnect(500)
	// TODO implement gracefull close of message channel
	close(messageChannel)

	// check for valid message format
	if currentState == internal.PowerOnSignal || currentState == internal.PowerOffSignal {
		return &currentState, nil
	}

	return nil, fmt.Errorf("unexpected state %s found, expected where %s or %s", currentState, internal.PowerOnSignal, internal.PowerOffSignal)

}

// SetupWithManager sets up the controller with the Manager.
func (r *PoweroutletReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&personaliotv1alpha1.Poweroutlet{}).
		Complete(r)
}
