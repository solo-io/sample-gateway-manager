package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	reasonOlderGatewayClassExists = "OlderGatewayClassExists"
	msgOlderGatewayClassExists    = "An older GatewayClass with the same controller exists"
)

func (p *Processor) updateStatus(ctx context.Context, obj client.Object) error {
	// Determine the object type to update.
	switch obj {
	case obj.(*gwapiv1b1.GatewayClass):
		gc, ok := obj.(*gwapiv1b1.GatewayClass)
		if !ok {
			return fmt.Errorf("failed to cast object: %v", obj)
		}
		if err := p.updateGatewayClassStatus(ctx, gc); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown object kind: %v", obj.GetObjectKind())
	}

	return nil
}

func (p *Processor) updateGatewayClassStatus(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	// The gatewayclass was already deleted/finalized and there are stale queue entries.
	accepted := p.ObjectStore.gatewayclasses.accepted()
	if accepted.Name == gc.Name {
		//if !status.IsEqual(accepted, gc) {
		p.setAcceptedStatus(ctx, gc)
		//}
	}

	/*all := p.ObjectStore.gatewayclasses.all()
	for _, class := range all {
		if class.Name == accepted.Name ||
			class.Name != gc.Name {
			continue
		}
		if !status.IsEqual(class, gc) {
			p.setNotAcceptedStatus(ctx, gc)
		}
	}*/

	// No Status updated needed.
	return nil
}

func (p *Processor) setAcceptedStatus(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	//status.ComputeGatewayClassAcceptedCondition(gc, true, )
	copy := gc.DeepCopy()
	acceptedCond := metav1.Condition{
		Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: gc.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gwapiv1b1.GatewayClassReasonAccepted),
		Message:            "gatewayclass is accepted",
	}
	setCondition(acceptedCond, gc)
	return p.Status().Patch(ctx, gc, client.MergeFrom(copy))
}

func (p *Processor) setNotAcceptedStatus(ctx context.Context, gc *gwapiv1b1.GatewayClass) error {
	copy := gc.DeepCopy()
	acceptedCond := metav1.Condition{
		Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
		Status:             metav1.ConditionFalse,
		ObservedGeneration: gc.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reasonOlderGatewayClassExists,
		Message:            msgOlderGatewayClassExists,
	}
	setCondition(acceptedCond, gc)
	return p.Status().Patch(ctx, gc, client.MergeFrom(copy))
}

func setCondition(condition metav1.Condition, gc *gwapiv1b1.GatewayClass) {
	newConds := make([]metav1.Condition, 0, len(gc.Status.Conditions))

	for i := 0; i < len(gc.Status.Conditions); i++ {
		if gc.Status.Conditions[i].Type != condition.Type {
			newConds = append(newConds, gc.Status.Conditions[i])
		}
	}

	newConds = append(newConds, condition)
	gc.Status.Conditions = newConds
}
