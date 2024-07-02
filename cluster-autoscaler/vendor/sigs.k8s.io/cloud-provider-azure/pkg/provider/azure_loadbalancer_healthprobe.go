/*
Copyright 2023 The Kubernetes Authors.

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

package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

func (az *Cloud) buildClusterServiceSharedProbe() *network.Probe {
	return &network.Probe{
		Name: pointer.String(consts.SharedProbeName),
		ProbePropertiesFormat: &network.ProbePropertiesFormat{
			Protocol:          network.ProbeProtocolHTTP,
			Port:              pointer.Int32(az.ClusterServiceSharedLoadBalancerHealthProbePort),
			RequestPath:       pointer.String(az.ClusterServiceSharedLoadBalancerHealthProbePath),
			IntervalInSeconds: pointer.Int32(consts.HealthProbeDefaultProbeInterval),
			ProbeThreshold:    pointer.Int32(consts.HealthProbeDefaultNumOfProbe),
		},
	}
}

// buildHealthProbeRulesForPort
// for following sku: basic loadbalancer vs standard load balancer
// for following protocols: TCP HTTP HTTPS(SLB only)
// return nil if no new probe is added
func (az *Cloud) buildHealthProbeRulesForPort(serviceManifest *v1.Service, port v1.ServicePort, lbrule string, healthCheckNodePortProbe *network.Probe, useSharedProbe bool) (*network.Probe, error) {
	if useSharedProbe {
		klog.V(4).Infof("skip creating health probe for port %d because the shared probe is used", port.Port)
		return nil, nil
	}

	if port.Protocol == v1.ProtocolUDP || port.Protocol == v1.ProtocolSCTP {
		return nil, nil
	}
	// protocol should be tcp, because sctp is handled in outer loop

	properties := &network.ProbePropertiesFormat{}
	var err error

	// order - Specific Override
	// port_ annotation
	// global annotation
	// Lookup or Override Health Probe Port

	probePort, err := consts.GetHealthProbeConfigOfPortFromK8sSvcAnnotation(serviceManifest.Annotations, port.Port, consts.HealthProbeParamsPort, func(s *string) error {
		if s == nil {
			return nil
		}
		//not a integer
		for _, item := range serviceManifest.Spec.Ports {
			if strings.EqualFold(item.Name, *s) {
				//found the port
				return nil
			}
		}
		//nolint:gosec
		port, err := strconv.Atoi(*s)
		if err != nil {
			return fmt.Errorf("port %s not found in service", *s)
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("port %d is out of range", port)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.BuildHealthProbeAnnotationKeyForPort(port.Port, consts.HealthProbeParamsPort), err)
	}

	if probePort != nil {
		//nolint:gosec
		port, err := strconv.ParseInt(*probePort, 10, 32)
		if err != nil {
			//not a integer
			for _, item := range serviceManifest.Spec.Ports {
				if strings.EqualFold(item.Name, *probePort) {
					//found the port
					properties.Port = pointer.Int32(item.NodePort)
				}
			}
		} else {
			// Not need to verify probePort is in correct range again.
			var found bool
			for _, item := range serviceManifest.Spec.Ports {
				//nolint:gosec
				if item.Port == int32(port) {
					//found the port
					properties.Port = pointer.Int32(item.NodePort)
					found = true
					break
				}
			}
			if !found {
				//nolint:gosec
				properties.Port = pointer.Int32(int32(port))
			}
		}
	} else if healthCheckNodePortProbe != nil {
		return nil, nil
	} else {
		properties.Port = &port.NodePort
	}
	// Select Protocol
	//
	var protocol *string

	// 1. Look up port-specific override
	protocol, err = consts.GetHealthProbeConfigOfPortFromK8sSvcAnnotation(serviceManifest.Annotations, port.Port, consts.HealthProbeParamsProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.BuildHealthProbeAnnotationKeyForPort(port.Port, consts.HealthProbeParamsProtocol), err)
	}

	// 2. If not specified, look up from AppProtocol
	// Note - this order is to remain compatible with previous versions
	if protocol == nil {
		protocol = port.AppProtocol
	}

	// 3. If protocol is still nil, check the global annotation
	if protocol == nil {
		protocol, err = consts.GetAttributeValueInSvcAnnotation(serviceManifest.Annotations, consts.ServiceAnnotationLoadBalancerHealthProbeProtocol)
		if err != nil {
			return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.ServiceAnnotationLoadBalancerHealthProbeProtocol, err)
		}
	}

	// 4. Finally, if protocol is still nil, default to TCP
	if protocol == nil {
		protocol = pointer.String(string(network.ProtocolTCP))
	}

	*protocol = strings.TrimSpace(*protocol)
	switch {
	case strings.EqualFold(*protocol, string(network.ProtocolTCP)):
		properties.Protocol = network.ProbeProtocolTCP
	case strings.EqualFold(*protocol, string(network.ProtocolHTTPS)):
		//HTTPS probe is only supported in standard loadbalancer
		//For backward compatibility,when unsupported protocol is used, fall back to tcp protocol in basic lb mode instead
		if !az.useStandardLoadBalancer() {
			properties.Protocol = network.ProbeProtocolTCP
		} else {
			properties.Protocol = network.ProbeProtocolHTTPS
		}
	case strings.EqualFold(*protocol, string(network.ProtocolHTTP)):
		properties.Protocol = network.ProbeProtocolHTTP
	default:
		//For backward compatibility,when unsupported protocol is used, fall back to tcp protocol in basic lb mode instead
		properties.Protocol = network.ProbeProtocolTCP
	}

	// Select request path
	if strings.EqualFold(string(properties.Protocol), string(network.ProtocolHTTPS)) || strings.EqualFold(string(properties.Protocol), string(network.ProtocolHTTP)) {
		// get request path ,only used with http/https probe
		path, err := consts.GetHealthProbeConfigOfPortFromK8sSvcAnnotation(serviceManifest.Annotations, port.Port, consts.HealthProbeParamsRequestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.BuildHealthProbeAnnotationKeyForPort(port.Port, consts.HealthProbeParamsRequestPath), err)
		}
		if path == nil {
			if path, err = consts.GetAttributeValueInSvcAnnotation(serviceManifest.Annotations, consts.ServiceAnnotationLoadBalancerHealthProbeRequestPath); err != nil {
				return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.ServiceAnnotationLoadBalancerHealthProbeRequestPath, err)
			}
		}
		if path == nil {
			path = pointer.String(consts.HealthProbeDefaultRequestPath)
		}
		properties.RequestPath = path
	}

	properties.IntervalInSeconds, properties.ProbeThreshold, err = az.getHealthProbeConfigProbeIntervalAndNumOfProbe(serviceManifest, port.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to parse health probe config for port %d: %w", port.Port, err)
	}
	probe := &network.Probe{
		Name:                  &lbrule,
		ProbePropertiesFormat: properties,
	}
	return probe, nil
}

// getHealthProbeConfigProbeIntervalAndNumOfProbe
func (az *Cloud) getHealthProbeConfigProbeIntervalAndNumOfProbe(serviceManifest *v1.Service, port int32) (*int32, *int32, error) {

	numberOfProbes, err := az.getHealthProbeConfigNumOfProbe(serviceManifest, port)
	if err != nil {
		return nil, nil, err
	}

	probeInterval, err := az.getHealthProbeConfigProbeInterval(serviceManifest, port)
	if err != nil {
		return nil, nil, err
	}
	// total probe should be less than 120 seconds ref: https://docs.microsoft.com/en-us/rest/api/load-balancer/load-balancers/create-or-update#probe
	if (*probeInterval)*(*numberOfProbes) >= 120 {
		return nil, nil, fmt.Errorf("total probe should be less than 120, please adjust interval and number of probe accordingly")
	}
	return probeInterval, numberOfProbes, nil
}

// getHealthProbeConfigProbeInterval get probe interval in seconds
// minimum probe interval in seconds is 5. ref: https://docs.microsoft.com/en-us/rest/api/load-balancer/load-balancers/create-or-update#probe
// if probeInterval is not set, set it to default instead ref: https://docs.microsoft.com/en-us/rest/api/load-balancer/load-balancers/create-or-update#probe
func (*Cloud) getHealthProbeConfigProbeInterval(serviceManifest *v1.Service, port int32) (*int32, error) {
	var probeIntervalValidator = func(val *int32) error {
		const (
			MinimumProbeIntervalInSecond = 5
		)
		if *val < 5 {
			return fmt.Errorf("the minimum value of %s is %d", consts.HealthProbeParamsProbeInterval, MinimumProbeIntervalInSecond)
		}
		return nil
	}
	probeInterval, err := consts.GetInt32HealthProbeConfigOfPortFromK8sSvcAnnotation(serviceManifest.Annotations, port, consts.HealthProbeParamsProbeInterval, probeIntervalValidator)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotation %s:%w", consts.BuildHealthProbeAnnotationKeyForPort(port, consts.HealthProbeParamsProbeInterval), err)
	}
	if probeInterval == nil {
		if probeInterval, err = consts.Getint32ValueFromK8sSvcAnnotation(serviceManifest.Annotations, consts.ServiceAnnotationLoadBalancerHealthProbeInterval, probeIntervalValidator); err != nil {
			return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.ServiceAnnotationLoadBalancerHealthProbeInterval, err)
		}
	}

	if probeInterval == nil {
		probeInterval = pointer.Int32(consts.HealthProbeDefaultProbeInterval)
	}
	return probeInterval, nil
}

// getHealthProbeConfigNumOfProbe get number of probes
// minimum number of unhealthy responses is 2. ref: https://docs.microsoft.com/en-us/rest/api/load-balancer/load-balancers/create-or-update#probe
// if numberOfProbes is not set, set it to default instead ref: https://docs.microsoft.com/en-us/rest/api/load-balancer/load-balancers/create-or-update#probe
func (*Cloud) getHealthProbeConfigNumOfProbe(serviceManifest *v1.Service, port int32) (*int32, error) {
	var numOfProbeValidator = func(val *int32) error {
		const (
			MinimumNumOfProbe = 2
		)
		if *val < MinimumNumOfProbe {
			return fmt.Errorf("the minimum value of %s is %d", consts.HealthProbeParamsNumOfProbe, MinimumNumOfProbe)
		}
		return nil
	}
	numberOfProbes, err := consts.GetInt32HealthProbeConfigOfPortFromK8sSvcAnnotation(serviceManifest.Annotations, port, consts.HealthProbeParamsNumOfProbe, numOfProbeValidator)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.BuildHealthProbeAnnotationKeyForPort(port, consts.HealthProbeParamsNumOfProbe), err)
	}
	if numberOfProbes == nil {
		if numberOfProbes, err = consts.Getint32ValueFromK8sSvcAnnotation(serviceManifest.Annotations, consts.ServiceAnnotationLoadBalancerHealthProbeNumOfProbe, numOfProbeValidator); err != nil {
			return nil, fmt.Errorf("failed to parse annotation %s: %w", consts.ServiceAnnotationLoadBalancerHealthProbeNumOfProbe, err)
		}
	}

	if numberOfProbes == nil {
		numberOfProbes = pointer.Int32(consts.HealthProbeDefaultNumOfProbe)
	}
	return numberOfProbes, nil
}

func findProbe(probes []network.Probe, probe network.Probe) bool {
	for _, existingProbe := range probes {
		if strings.EqualFold(pointer.StringDeref(existingProbe.Name, ""), pointer.StringDeref(probe.Name, "")) &&
			pointer.Int32Deref(existingProbe.Port, 0) == pointer.Int32Deref(probe.Port, 0) &&
			strings.EqualFold(string(existingProbe.Protocol), string(probe.Protocol)) &&
			strings.EqualFold(pointer.StringDeref(existingProbe.RequestPath, ""), pointer.StringDeref(probe.RequestPath, "")) &&
			pointer.Int32Deref(existingProbe.IntervalInSeconds, 0) == pointer.Int32Deref(probe.IntervalInSeconds, 0) &&
			pointer.Int32Deref(existingProbe.ProbeThreshold, 0) == pointer.Int32Deref(probe.ProbeThreshold, 0) {
			return true
		}
	}
	return false
}

// keepSharedProbe ensures the shared probe will not be removed if there are more than 1 service referencing it.
func (az *Cloud) keepSharedProbe(
	service *v1.Service,
	lb network.LoadBalancer,
	expectedProbes []network.Probe,
	wantLB bool,
) ([]network.Probe, error) {
	var shouldConsiderRemoveSharedProbe bool
	if !wantLB {
		shouldConsiderRemoveSharedProbe = true
	}

	if lb.LoadBalancerPropertiesFormat != nil && lb.Probes != nil {
		for _, probe := range *lb.Probes {
			if strings.EqualFold(pointer.StringDeref(probe.Name, ""), consts.SharedProbeName) {
				if !az.useSharedLoadBalancerHealthProbeMode() {
					shouldConsiderRemoveSharedProbe = true
				}
				if probe.ProbePropertiesFormat != nil && probe.LoadBalancingRules != nil {
					for _, rule := range *probe.LoadBalancingRules {
						ruleName, err := getLastSegment(*rule.ID, "/")
						if err != nil {
							klog.Errorf("failed to parse load balancing rule name %s attached to health probe %s", *rule.ID, *probe.ID)
							return []network.Probe{}, err
						}
						if !az.serviceOwnsRule(service, ruleName) && shouldConsiderRemoveSharedProbe {
							klog.V(4).Infof("there are load balancing rule %s of another service referencing the health probe %s, so the health probe should not be removed", *rule.ID, *probe.ID)
							sharedProbe := az.buildClusterServiceSharedProbe()
							expectedProbes = append(expectedProbes, *sharedProbe)
							return expectedProbes, nil
						}
					}
				}
			}
		}
	}
	return expectedProbes, nil
}
