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
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"solo.io/sample-gateway-manager/internal/gatewayapi"

	"github.com/go-logr/logr"
	//corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"solo.io/sample-gateway-manager/internal/model"
)

//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways/finalizers,verbs=update

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get

//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints/status,verbs=get

// GatewayReconciler reconciles a Gateway object
type GatewayReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *model.ManagerConfig
	Log    logr.Logger

	ProcessorChan chan event.GenericEvent
	ObjectStore   *ObjectStore
}

func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = log.FromContext(context.Background()).WithName("gateway reconciler")

	return ctrl.NewControllerManagedBy(mgr).
		For(&gwapiv1b1.Gateway{},
			builder.WithPredicates(predicate.NewPredicateFuncs(r.gatewayHasMatchingGatewayClass)),
		).
		/*Watches(
			&source.Kind{Type: &corev1.Service{}},
			handler.EnqueueRequestsFromMapFunc(mapServiceToGateway),
		).
		Watches(
			&source.Kind{Type: &gwapiv1b1.GatewayClass{}},
			handler.EnqueueRequestsFromMapFunc(r.mapGatewayClassToGateway),
		).*/
		Complete(r)
}

func (r *GatewayReconciler) gatewayHasMatchingGatewayClass(obj client.Object) bool {
	gw, ok := obj.(*gwapiv1b1.Gateway)
	if !ok {
		r.Log.Error(fmt.Errorf("unexpected object type"), "expected", "Gateway", "found", obj.GetObjectKind())
		return false
	}

	gc := &gwapiv1b1.GatewayClass{}
	if err := r.Client.Get(context.Background(), client.ObjectKey{Name: string(gw.Spec.GatewayClassName)}, gc); err != nil {
		r.Log.Error(err, "failed to get gatewayclass", "name", gw.Spec.GatewayClassName)
		return false
	}

	return string(gc.Spec.ControllerName) == r.Config.ControllerName
}

func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling request", "namespace", req.Namespace, "name", req.Name)

	gw := new(gwapiv1b1.Gateway)
	if err := r.Client.Get(ctx, req.NamespacedName, gw); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("reconciled object no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Process the gatewayclass if it doesn't exist or differs from the internal store.
	gc, ok := r.ObjectStore.gatewayclasses.matched[gatewayapi.ObjectNameToStr(gw.Spec.GatewayClassName)]
	if !ok {
		gc = gwapiv1b1.GatewayClass{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: string(gw.Spec.GatewayClassName)}, &gc); err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
	}

	if string(gc.Spec.ControllerName) != r.Config.ControllerName {
		return ctrl.Result{}, nil
	}

	current, ok := r.ObjectStore.gateways[req.NamespacedName]

	// Process the gateway if it doesn't exist or differs from the internal store.
	if !ok || !reflect.DeepEqual(gw, current) {
		r.ObjectStore.mu.Lock()
		defer r.ObjectStore.mu.Unlock()
		r.ObjectStore.gateways[req.NamespacedName] = *gw
		update := event.GenericEvent{Object: gw}
		r.ProcessorChan <- update
	}

	/*gatewayReadyStatus, gatewayReadyStatusIsSet := isGatewayReady(gw)
	oldGateway := gw.DeepCopy()
	initGatewayStatus(gw)
	factorizeStatus(gw, oldGateway)
	if !gatewayReadyStatusIsSet {
		return ctrl.Result{}, r.Status().Patch(ctx, gw, client.MergeFrom(oldGateway))
	}

	Log.Info("checking for Service for Gateway")
	svc, err := r.getServiceForGateway(ctx, gw)
	if err != nil {
		return ctrl.Result{}, err
	}
	if svc == nil {
		// if the ready status is not set, or the gateway is marked as ready, mark it as not ready
		if gatewayReadyStatus {
			return ctrl.Result{}, r.Status().Patch(ctx, gw, client.MergeFrom(oldGateway)) // status patch will requeue gateway
		}
		Log.Info("creating Service for Gateway")
		return ctrl.Result{}, r.createServiceForGateway(ctx, gw) // service creation will requeue gateway
	}

	Log.Info("checking Service configuration")
	needsUpdate, err := r.ensureServiceConfiguration(ctx, svc, gw)
	// in both cases when the service does not exist or an error has been triggered, the Gateway
	// must be not ready. This OR condition is redundant, as (needsUpdate == true AND err == nil)
	// should never happen, but useful to highlight the purpose.
	if err != nil {
		return ctrl.Result{}, err
	}
	if needsUpdate {
		// if the ready status is not set, or the gateway is marked as ready, mark it as not ready
		if gatewayReadyStatus {
			return ctrl.Result{}, r.Status().Patch(ctx, gw, client.MergeFrom(oldGateway)) // status patch will requeue gateway
		}
		return ctrl.Result{}, r.Client.Update(ctx, svc)
	}

	Log.Info("checking Service status", "namespace", svc.Namespace, "name", svc.Name)
	switch t := svc.Spec.Type; t {
	case corev1.ServiceTypeLoadBalancer:
		if svc.Spec.ClusterIP == "" || len(svc.Status.LoadBalancer.Ingress) < 1 {
			// if the ready status is not set, or the gateway is marked as ready, mark it as not ready
			if gatewayReadyStatus {
				return ctrl.Result{}, r.Status().Patch(ctx, gw, client.MergeFrom(oldGateway)) // status patch will requeue gateway
			}
			Log.Info("waiting for Service to be ready")
			return ctrl.Result{Requeue: true}, nil
		}
	default:
		// if the ready status is not set, or the gateway is marked as ready, mark it as not ready
		if gatewayReadyStatus {
			return ctrl.Result{}, r.Status().Patch(ctx, gw, client.MergeFrom(oldGateway)) // status patch will requeue gateway
		}
		return ctrl.Result{}, fmt.Errorf("found unsupported Service type: %s (only LoadBalancer type is currently supported)", t)
	}

	// hack for metallb - https://github.com/metallb/metallb/issues/1640
	// no need to enforce the gateway status here, as this endpoint is not reconciled by the controller
	// and no reconciliation loop is triggered upon its change or deletion.
	created, err := r.hackEnsureEndpoints(ctx, svc)
	if err != nil {
		return ctrl.Result{}, err
	}
	if created {
		return ctrl.Result{Requeue: true}, nil
	}

	Log.Info("Service is ready, updating Gateway")
	updateGatewayStatus(ctx, gw, svc)
	factorizeStatus(gw, oldGateway)*/

	r.Log.Info("reconciled request", "namespace", req.Namespace, "name", req.Name)

	return ctrl.Result{}, nil
}
