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

package model

import gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

type ManagerConfig struct {
	ControllerName string
}

type ManagedClasses struct {
	// Matched stores all GatewayClass objects with a controllerName.
	Matched []*gwapiv1b1.GatewayClass

	// Oldest stores the first GatewayClass encountered with matching
	// controllerName.
	Oldest *gwapiv1b1.GatewayClass
}
