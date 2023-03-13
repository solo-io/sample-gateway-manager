/*
Copyright 2023.

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

package kubernetes

import (
	"context"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"solo.io/sample-gateway-manager/internal/model"
)

//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses/finalizers,verbs=update

// GatewayClassReconciler reconciles a GatewayClass object
type GatewayClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *model.ManagerConfig
	Log    logr.Logger

	ProcessorChan chan event.GenericEvent
	ObjectStore   *ObjectStore
}

func (r *GatewayClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = log.FromContext(context.Background()).WithName("gatewayclass reconciler")

	return ctrl.NewControllerManagedBy(mgr).
		For(&gwapiv1b1.GatewayClass{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			gc, ok := obj.(*gwapiv1b1.GatewayClass)
			if !ok {
				return false
			}
			return string(gc.Spec.ControllerName) == r.Config.ControllerName
		})).
		Complete(r)
}

func (r *GatewayClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling request", "name", req.Name)

	gc := new(gwapiv1b1.GatewayClass)
	if err := r.Client.Get(ctx, req.NamespacedName, gc); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("object no longer exists")
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get gatewayclass", "name", req.Name)
		return ctrl.Result{}, err
	}

	if string(gc.Spec.ControllerName) != r.Config.ControllerName {
		r.Log.Info("gatewayclass controller name doesn't match configuration; bypassing", "name", req.Name)
		return ctrl.Result{}, nil
	}

	// Process the gatewayclass.
	existing, ok := r.ObjectStore.gatewayclasses.matched[gc.Name]
	if !ok || !reflect.DeepEqual(gc, existing) {
		r.ObjectStore.mu.Lock()
		defer r.ObjectStore.mu.Unlock()
		r.ObjectStore.gatewayclasses.add(gc)
		update := event.GenericEvent{Object: gc}
		r.ProcessorChan <- update
	}

	r.Log.Info("reconciled request", "name", req.Name)

	return ctrl.Result{}, nil
}
