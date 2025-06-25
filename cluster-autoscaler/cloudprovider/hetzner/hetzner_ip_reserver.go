/*
Copyright 2019 The Kubernetes Authors.

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

package hetzner

import (
    "context"
    "fmt"
    "net"
    "sync"

    "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
    "k8s.io/klog/v2"
)

// ReservedIPLabelName is the label key for reserved IPs on Hetzner servers
const ReservedIPLabelName = "cluster-autoscaler/reserved-ip"

// ipReserver manages IP address reservations
type ipReserver struct {
    client         *hcloud.Client
    apiCallContext context.Context
    cachedServers  *serversCache
    reservedIPs    map[string]net.IP  // Uses string representation as key for quick lookups
    mutex          sync.RWMutex       // Protects reservedIPs map for thread safety
}

// newIPReserver creates a new IP reserver instance
func newIPReserver(ctx context.Context, client *hcloud.Client, cache *serversCache) *ipReserver {
    if client == nil {
        klog.Fatalf("Failed to create ipReserver: client is nil")
    }
    if ctx == nil {
        klog.Fatalf("Failed to create ipReserver: context is nil")
    }
    if cache == nil {
        klog.Fatalf("Failed to create ipReserver: serversCache is nil")
    }
    return &ipReserver{
        client:         client,
        apiCallContext: ctx,
        cachedServers:  cache,
        reservedIPs:    make(map[string]net.IP),
    }
}

// getReservedIPLabelName returns the label name for reserved IPs
func (r *ipReserver) getReservedIPLabelName() string {
    return ReservedIPLabelName
}

// getReservedIPs returns a map of all currently reserved IPs with their string representation as key
func (r *ipReserver) getReservedIPs() map[string]net.IP {
    serverIPs, err := r.getReservedIPsFromServers()
    if err != nil {
        klog.Errorf("Failed to get reserved IPs from servers: %v", err)
        serverIPs = []net.IP{} // Fallback to empty slice if error occurs
    }

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    // Create result map with capacity for all IPs
    result := make(map[string]net.IP, len(r.reservedIPs)+len(serverIPs))

    // Add all IPs from local storage
    for ipStr, ip := range r.reservedIPs {
        result[ipStr] = ip
    }

    // Add server IPs if not already in result
    for _, ip := range serverIPs {
        if ip != nil {
            ipStr := ip.String()
            if _, exists := result[ipStr]; !exists {
                result[ipStr] = ip
            }
        }
    }

    return result
}

// addReservedIP adds an IP to the list of reserved IPs
func (r *ipReserver) addReservedIP(ip net.IP) {
    if ip == nil {
        klog.Warning("Attempted to add a nil IP to reserved IPs")
        return
    }

    r.mutex.Lock()
    defer r.mutex.Unlock()

    // Store a copy of the IP to prevent modification
    r.reservedIPs[ip.String()] = cloneIP(ip)
}

// removeReservedIP removes an IP from the list of reserved IPs
func (r *ipReserver) removeReservedIP(ip net.IP) {
    if ip == nil {
        klog.Warning("Attempted to remove a nil IP from reserved IPs")
        return
    }

    r.mutex.Lock()
    defer r.mutex.Unlock()

    if _, exists := r.reservedIPs[ip.String()]; !exists {
        klog.Warningf("Attempted to remove an IP that is not reserved: %s", ip.String())
        return
    }

    delete(r.reservedIPs, ip.String())
}

// getReservedIPsFromServers retrieves all IPs that are reserved on servers
func (r *ipReserver) getReservedIPsFromServers() ([]net.IP, error) {
    servers, err := r.cachedServers.getAllServers()
    if err != nil {
        return nil, fmt.Errorf("failed to get servers: %w", err)
    }

    ips := []net.IP{}
    for _, server := range servers {
        if server == nil {
            klog.Warning("Encountered a nil server while retrieving reserved IPs")
            continue
        }

        // Check for IPs in labels
        if ip, exists := r.getReservedIPFromLabel(server); exists {
            ips = append(ips, ip)
        }

        // Check for IPs in private networks
        for _, privNet := range server.PrivateNet {
            if !privNet.IP.IsUnspecified() {
                ips = append(ips, cloneIP(privNet.IP))
            }
        }
    }
    return ips, nil
}

// getReservedIPFromLabel extracts the reserved IP from server label
func (r *ipReserver) getReservedIPFromLabel(server *hcloud.Server) (net.IP, bool) {
    if server == nil || server.Labels == nil {
        klog.Warning("Attempted to retrieve reserved IP from a nil server or server with nil labels")
        return nil, false
    }

    ipLabelValue, exists := server.Labels[r.getReservedIPLabelName()]
    if !exists {
        return nil, false
    }

    parsedIP := net.ParseIP(ipLabelValue)
    if parsedIP == nil {
        klog.Warningf("Invalid reserved IP label value '%s' for server %s", ipLabelValue, server.Name)
        return nil, false
    }

    return parsedIP, true
}

// reserveNewIP reserves a new IP from the given subnet
func (r *ipReserver) reserveNewIP(subnet *net.IPNet) (net.IP, error) {
    if subnet == nil {
        return nil, fmt.Errorf("subnet cannot be nil")
    }

    // Get all currently used IPs - already as map for efficient lookup
    reservedIPs := r.getReservedIPs()

    // Find first available IP in subnet (skipping network and broadcast addresses)
    ip := cloneIP(subnet.IP)
    broadcast := getBroadcastAddress(subnet)

    for subnet.Contains(ip) {
        // Skip network and broadcast addresses
        if ip.Equal(subnet.IP) || ip.Equal(broadcast) {
            incrementIP(ip)
            continue
        }

        ipStr := ip.String()
        if _, exists := reservedIPs[ipStr]; !exists {
            reserved := cloneIP(ip)
            r.addReservedIP(reserved)
            return reserved, nil
        }

        incrementIP(ip)
    }

    return nil, fmt.Errorf("no free IP available in subnet %s", subnet.String())
}

// cloneIP creates a copy of an IP address
func cloneIP(ip net.IP) net.IP {
    if ip == nil {
        klog.Warning("Attempted to clone a nil IP")
        return nil
    }
    clone := make(net.IP, len(ip))
    copy(clone, ip)
    return clone
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
    for j := len(ip)-1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}

// getBroadcastAddress returns the broadcast address for a subnet
func getBroadcastAddress(subnet *net.IPNet) net.IP {
    if subnet == nil {
        klog.Warning("Attempted to get broadcast address for a nil subnet")
        return nil
    }

    broadcast := cloneIP(subnet.IP)
    for i := range broadcast {
        broadcast[i] |= ^subnet.Mask[i]
    }
    return broadcast
}