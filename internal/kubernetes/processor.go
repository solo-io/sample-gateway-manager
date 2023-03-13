package kubernetes

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"solo.io/sample-gateway-manager/internal/utils/slice"

	"solo.io/sample-gateway-manager/internal/model"
)

const (
	gatewayClassFinalizer = gwapiv1b1.GatewayClassFinalizerGatewaysExist
)

// Processor processes managed objects.
type Processor struct {
	client.Client
	Scheme      *runtime.Scheme
	Config      *model.ManagerConfig
	Log         logr.Logger
	UpdateChan  chan event.GenericEvent
	ObjectStore *ObjectStore
}

func (p *Processor) SetupWithManager(mgr ctrl.Manager) error {
	p.Log = log.FromContext(context.Background()).WithName("processor reconciler")

	return ctrl.NewControllerManagedBy(mgr).
		Named("processor").
		Watches(&source.Channel{Source: p.UpdateChan}, &handler.EnqueueRequestForObject{}).
		Complete(p)
}

func (p *Processor) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	p.Log.Info("reconciling request", "name", req.Name)

	p.Log.Info("request", "name", req.Name)
	p.Log.Info("object", "matched", p.ObjectStore)

	if _, ok := p.ObjectStore.gatewayclasses.matched[req.Name]; ok {
		if err := p.processGatewayClasses(ctx); err != nil {
			return ctrl.Result{}, err
		}
		p.Log.Info("processed gatewayclass", "name", req.Name)
	}

	if _, ok := p.ObjectStore.gateways[req.NamespacedName]; ok {
		if err := p.processGateways(ctx); err != nil {
			return ctrl.Result{}, err
		}
		p.Log.Info("processed gateway", "namespace", req.Namespace, "name", req.Name)
	}

	return ctrl.Result{}, nil
}

func (p *Processor) processGatewayClasses(ctx context.Context) error {
	for _, gc := range p.ObjectStore.gatewayclasses.matched {
		if !gc.DeletionTimestamp.IsZero() &&
			!slice.ContainsString(gc.Finalizers, gatewayClassFinalizer) {
			p.Log.Info("gatewayclass marked for deletion")
			// Delete the gatewayclass from the object store.
			delete(p.ObjectStore.gatewayclasses.matched, gc.Name)
			continue
		}
	}

	// Update status for all managed gatewayclasses.
	for _, class := range p.ObjectStore.gatewayclasses.all() {
		if err := p.updateGatewayClassStatus(ctx, &class); err != nil {
			delete(p.ObjectStore.gatewayclasses.matched, class.Name)
			return err
		}
	}

	return nil
}

func (p *Processor) processGateways(ctx context.Context) error {
	return nil
}
