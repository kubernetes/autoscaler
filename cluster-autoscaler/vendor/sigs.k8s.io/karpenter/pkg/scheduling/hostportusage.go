/*
Copyright The Kubernetes Authors.

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

package scheduling

import (
	"fmt"
	"net"

	"github.com/awslabs/operatorpkg/serrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go tool -modfile=../../go.tools.mod controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="."

// HostPortUsage tracks HostPort usage within a node. On a node, each <hostIP, hostPort, protocol> used by pods bound
// to the node must be unique. We need to track this to keep an accurate concept of what pods can potentially schedule
// together.
// +k8s:deepcopy-gen=true
type HostPortUsage struct {
	reserved map[types.NamespacedName][]HostPort
}

// +k8s:deepcopy-gen=true
type HostPort struct {
	IP       net.IP
	Port     int32
	Protocol v1.Protocol
}

func (p HostPort) String() string {
	return fmt.Sprintf("IP=%s Port=%d Proto=%s", p.IP, p.Port, p.Protocol)
}

func (p HostPort) Matches(rhs HostPort) bool {
	if p.Protocol != rhs.Protocol {
		return false
	}
	if p.Port != rhs.Port {
		return false
	}
	// If IPs are unequal, they don't match unless one is an unspecified address "0.0.0.0" or the IPv6 address "::".
	if !p.IP.Equal(rhs.IP) && !p.IP.IsUnspecified() && !rhs.IP.IsUnspecified() {
		return false
	}
	return true
}

func NewHostPortUsage() *HostPortUsage {
	return &HostPortUsage{
		reserved: map[types.NamespacedName][]HostPort{},
	}
}

// Add adds a port to the HostPortUsage
func (u *HostPortUsage) Add(usedBy *v1.Pod, ports []HostPort) {
	u.reserved[client.ObjectKeyFromObject(usedBy)] = ports
}

func (u *HostPortUsage) Conflicts(usedBy *v1.Pod, ports []HostPort) error {
	for _, newEntry := range ports {
		for podKey, entries := range u.reserved {
			for _, existing := range entries {
				if newEntry.Matches(existing) && podKey != client.ObjectKeyFromObject(usedBy) {
					return serrors.Wrap(fmt.Errorf("pod hostport conflicts with existing hostport configuration"), "pod-hostport-ip", newEntry.IP, "pod-hostport-port", newEntry.Port, "pod-hostport-protocol", newEntry.Protocol, "existing-hostport-ip", existing.IP, "existing-hostport-port", existing.Port, "existing-hostport-protocol", existing.Protocol)
				}
			}
		}
	}
	return nil
}

// DeletePod deletes all host port usage from the HostPortUsage that were created by the pod with the given name.
func (u *HostPortUsage) DeletePod(key types.NamespacedName) {
	delete(u.reserved, key)
}

func GetHostPorts(pod *v1.Pod) []HostPort {
	var usage []HostPort
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.HostPort == 0 {
				continue
			}
			// Per the K8s docs, "If you don't specify the hostIP and Protocol explicitly, Kubernetes will use 0.0.0.0
			// as the default hostIP and TCP as the default Protocol." In testing, and looking at the code the Protocol
			// is defaulted to TCP, but it leaves the IP empty.
			hostIP := p.HostIP
			if hostIP == "" {
				hostIP = "0.0.0.0"
			}
			usage = append(usage, HostPort{
				IP:       net.ParseIP(hostIP),
				Port:     p.HostPort,
				Protocol: p.Protocol,
			})
		}
	}
	return usage
}
