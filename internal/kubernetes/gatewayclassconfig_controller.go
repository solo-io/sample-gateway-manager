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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"solo.io/sample-gateway-manager/internal/model"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cfgv1a1 "solo.io/sample-gateway-manager/api/v1alpha1"
)

// GatewayClassConfigReconciler reconciles a GatewayClassConfig object
type GatewayClassConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *model.ManagerConfig
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=sample.io,resources=gatewayclassconfigs,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=sample.io,resources=gatewayclassconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sample.io,resources=gatewayclassconfigs/finalizers,verbs=update

// SetupWithManager sets up the controller with the Manager.
func (r *GatewayClassConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = log.FromContext(context.Background()).WithName("gatewayclassconfig reconciler")

	return ctrl.NewControllerManagedBy(mgr).
		For(&cfgv1a1.GatewayClassConfig{}).
		Complete(r)
}

func (r *GatewayClassConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling request", "namespace", req.Namespace, "name", req.Name)

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

	r.Log.Info("reconciled request", "namespace", req.Namespace, "name", req.Name)

	return ctrl.Result{}, nil
}
