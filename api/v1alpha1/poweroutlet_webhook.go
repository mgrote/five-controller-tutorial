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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/mgrote/personal-iot/internal"
)

// log is for logging in this package.
var poweroutletlog = logf.Log.WithName("poweroutlet-resource")

func (r *Poweroutlet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-personal-iot-frup-org-v1alpha1-poweroutlet,mutating=true,failurePolicy=fail,sideEffects=None,groups=personal-iot.frup.org,resources=poweroutlets,verbs=create;update,versions=v1alpha1,name=mpoweroutlet.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Poweroutlet{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Poweroutlet) Default() {
	poweroutletlog.Info("default", "name", r.Name)

	// TODO lecture ---> why is OFF set in any case (enum with default)
	if r.Spec.Switch == "" {
		r.Spec.Switch = internal.PowerOffSignal
	}
	// TODO lecture ---> code later in stead of switch
	if r.Spec.OutletName == "" {
		r.Spec.OutletName = "outlet" + rand.String(4)
	}
	if r.Name == "" {
		r.Name = r.Spec.OutletName
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-personal-iot-frup-org-v1alpha1-poweroutlet,mutating=false,failurePolicy=fail,sideEffects=None,groups=personal-iot.frup.org,resources=poweroutlets,verbs=create;update,versions=v1alpha1,name=vpoweroutlet.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Poweroutlet{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Poweroutlet) ValidateCreate() error {
	poweroutletlog.Info("validate create", "name", r.Name)

	if r.Spec.Switch == internal.PowerOnSignal {
		return fmt.Errorf("%s is an undesired state for a switch during create, please check your device and set Spec.Switch to %s", internal.PowerOnSignal, internal.PowerOffSignal)
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Poweroutlet) ValidateUpdate(old runtime.Object) error {
	poweroutletlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Poweroutlet) ValidateDelete() error {
	poweroutletlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
