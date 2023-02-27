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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/internal/mqttiot"
)

// PowerstripReconciler reconciles a Powerstrip object
type PowerstripReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	mqttSubscriber mqttiot.MQTTSubscriber
}

//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=powerstrips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=powerstrips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=personal-iot.frup.org,resources=powerstrips/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Powerstrip object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *PowerstripReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("PowerstripReconciler", req.NamespacedName)

	powerStrip := &personaliotv1alpha1.Powerstrip{}
	if err := r.Get(ctx, req.NamespacedName, powerStrip); err != nil {
		logger.Error(err, "unable to fetch power outlet")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	location, err := r.createLocationIfNotExists(ctx, powerStrip)
	if err != nil {
		return ctrl.Result{}, err
	}

	outlets, err := r.getOrCreateOutlets(ctx, powerStrip)
	if err != nil {
		return ctrl.Result{}, err
	}

	existingOutletNames, err := r.checkOutletReachability(outlets)
	if err != nil {
		return ctrl.Result{}, err
	}

	powerStrip.Status.Location = location.Name
	powerStrip.Status.Outlets = existingOutletNames
	if err := r.Status().Update(ctx, powerStrip); err != nil {
		logger.Error(err, "update PowerOutlet status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PowerstripReconciler) checkOutletReachability(outlets []*personaliotv1alpha1.Poweroutlet) ([]string, error) {

	if err := r.mqttSubscriber.Connect(); err != nil {
		return nil, err
	}

	defer r.mqttSubscriber.Disconnect(500)

	var existingOutletNames []string
	for _, outlet := range outlets {
		messageChannel := make(chan mqttiot.MQTTMessage)
		if err := r.mqttSubscriber.Subscribe(outlet.Spec.MQTTStatusTopik, 1, messageChannel); err != nil {
			return nil, err
		}

		incoming := <-messageChannel
		if len(incoming.Msg) > 1 {
			existingOutletNames = append(existingOutletNames, outlet.Name)
		}

		close(messageChannel)
	}
	return existingOutletNames, nil
}

func (r *PowerstripReconciler) getOrCreateOutlets(ctx context.Context, powerStrip *personaliotv1alpha1.Powerstrip) ([]*personaliotv1alpha1.Poweroutlet, error) {
	desiredOutlets := powerStrip.Spec.Outlets
	var existingOutlets []*personaliotv1alpha1.Poweroutlet
	for _, outlet := range desiredOutlets {
		err := r.Get(ctx, client.ObjectKey{Namespace: powerStrip.Namespace, Name: outlet.Spec.OutletName}, outlet)
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		if errors.IsNotFound(err) {
			outlet.Name = outlet.Spec.OutletName
			outlet.Namespace = powerStrip.Namespace
			err = r.Create(ctx, outlet)
			if err != nil {
				return nil, err
			}
		}
		existingOutlets = append(existingOutlets, outlet)
	}
	return existingOutlets, nil
}

func (r *PowerstripReconciler) createLocationIfNotExists(ctx context.Context, powerStrip *personaliotv1alpha1.Powerstrip) (*personaliotv1alpha1.Location, error) {
	location := &personaliotv1alpha1.Location{}
	err := r.Get(ctx, client.ObjectKey{Namespace: powerStrip.Namespace, Name: powerStrip.Spec.LocationName}, location)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}
	if errors.IsNotFound(err) {
		location.Namespace = powerStrip.Namespace
		location.Name = powerStrip.Spec.LocationName
		err = r.Create(ctx, location)
		if err != nil {
			return nil, err
		}
	}
	return location, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PowerstripReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&personaliotv1alpha1.Powerstrip{}).
		Complete(r)
}
