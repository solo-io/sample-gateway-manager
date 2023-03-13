package kubernetes

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type ObjectStore struct {
	mu sync.Mutex
	// Map for storing managed gatewayclasses.
	gatewayclasses *managedClasses
	// Map for storing managed gateways.
	gateways map[types.NamespacedName]gwapiv1b1.Gateway
}

type managedClasses struct {
	matched map[string]gwapiv1b1.GatewayClass
	oldest  *gwapiv1b1.GatewayClass
}

func NewObjectStore() *ObjectStore {
	return &ObjectStore{
		gatewayclasses: &managedClasses{
			matched: make(map[string]gwapiv1b1.GatewayClass),
			oldest:  new(gwapiv1b1.GatewayClass),
		},
		gateways: map[types.NamespacedName]gwapiv1b1.Gateway{},
	}
}

func (mc *managedClasses) add(gc *gwapiv1b1.GatewayClass) {
	mc.matched[gc.Name] = *gc

	switch {
	case gc.CreationTimestamp.Time.Before(mc.oldest.CreationTimestamp.Time):
		mc.oldest = gc
	case gc.CreationTimestamp.Time.Equal(mc.oldest.CreationTimestamp.Time) && gc.Name < mc.oldest.Name:
		// The first one in alphabetical order is considered oldest/accepted.
		mc.oldest = gc
	}
}

func (mc *managedClasses) remove(gc *gwapiv1b1.GatewayClass) {
	// First remove gc from matched.
	delete(mc.matched, gc.Name)

	// If gc is the oldest gatewayclass, remove it.
	if mc.oldest != nil && mc.oldest.Name == gc.Name {
		mc.oldest = new(gwapiv1b1.GatewayClass)
		// Set the new oldest from the matched gatewayclasses.
		for _, matchedGC := range mc.matched {
			if matchedGC.CreationTimestamp.Time.Before(mc.oldest.CreationTimestamp.Time) ||
				(matchedGC.CreationTimestamp.Time.Equal(mc.oldest.CreationTimestamp.Time) &&
					matchedGC.Name < mc.oldest.Name) {
				mc.oldest = &matchedGC
				return
			}
		}
	}
}

func (mc *managedClasses) accepted() *gwapiv1b1.GatewayClass {
	return mc.oldest
}

func (mc *managedClasses) notAccepted() []gwapiv1b1.GatewayClass {
	var res []gwapiv1b1.GatewayClass
	for _, gc := range mc.matched {
		// Skip the oldest one since it will be accepted.
		if gc.Name != mc.oldest.Name {
			res = append(res, gc)
		}
	}

	return res
}

func (mc *managedClasses) all() []gwapiv1b1.GatewayClass {
	res := []gwapiv1b1.GatewayClass{*mc.accepted()}
	na := mc.notAccepted()
	res = append(res, na...)

	return res
}
