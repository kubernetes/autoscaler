/*
Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.

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

// Package machine is the internal version of the API.
package machine

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName of all machine-controller-manager objects
const GroupName = "machine.sapcloud.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder needs to be exported as `SchemeBuilder` so
	// the code-generation can find it.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is exposed for API installation
	AddToScheme        = SchemeBuilder.AddToScheme
	localSchemeBuilder = &SchemeBuilder
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&OpenStackMachineClass{},
		&OpenStackMachineClassList{},

		&AWSMachineClass{},
		&AWSMachineClassList{},

		&AzureMachineClass{},
		&AzureMachineClassList{},

		&GCPMachineClass{},
		&GCPMachineClassList{},

		&AlicloudMachineClass{},
		&AlicloudMachineClassList{},

		&PacketMachineClass{},
		&PacketMachineClassList{},

		&MachineClass{},
		&MachineClassList{},

		&Machine{},
		&MachineList{},

		&MachineSet{},
		&MachineSetList{},

		&MachineDeployment{},
		&MachineDeploymentList{},
	)
	return nil
}
