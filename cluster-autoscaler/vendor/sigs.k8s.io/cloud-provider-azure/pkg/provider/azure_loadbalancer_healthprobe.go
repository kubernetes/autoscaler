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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

func (az *Cloud) buildClusterServiceSharedProbe() *armnetwork.Probe {
	return &armnetwork.Probe{
		Name: ptr.To(consts.SharedProbeName),
		Properties: &armnetwork.ProbePropertiesFormat{
			Protocol:          to.Ptr(armnetwork.ProbeProtocolHTTP),
			Port:              ptr.To(az.ClusterServiceSharedLoadBalancerHealthProbePort),
			RequestPath:       ptr.To(az.ClusterServiceSharedLoadBalancerHealthProbePath),
			IntervalInSeconds: ptr.To(consts.HealthProbeDefaultProbeInterval),
			ProbeThreshold:    ptr.To(consts.HealthProbeDefaultNumOfProbe),
		},
	}
}

// buildHealthProbeRulesForPort
// for following SKU: basic loadbalancer vs standard load balancer
// for following protocols: TCP HTTP HTTPS(SLB only)
// return nil if no new probe is added
func (az *Cloud) buildHealthProbeRulesForPort(serviceManifest *v1.Service, port v1.ServicePort, lbrule string, healthCheckNodePortProbe *armnetwork.Probe, useSharedProbe bool) (*armnetwork.Probe, error) {
	logger := klog.Background().WithName("buildHealthProbeRulesForPort")
	if useSharedProbe {
		logger.V(4).Info("skip creating health probe for port because the shared probe is used", "port", port.Port)
		return nil, nil
	}

	if port.Protocol == v1.ProtocolUDP || port.Protocol == v1.ProtocolSCTP {
		return nil, nil
	}
	// protocol should be tcp, because sctp is handled in outer loop

	properties := &armnetwork.ProbePropertiesFormat{}
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
					properties.Port = ptr.To(item.NodePort)
				}
			}
		} else {
			// Not need to verify probePort is in correct range again.
			var found bool
			for _, item := range serviceManifest.Spec.Ports {
				//nolint:gosec
				if item.Port == int32(port) {
					//found the port
					properties.Port = ptr.To(item.NodePort)
					found = true
					break
				}
			}
			if !found {
				//nolint:gosec
				properties.Port = ptr.To(int32(port))
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
		protocol = ptr.To(string(armnetwork.ProtocolTCP))
	}

	*protocol = strings.TrimSpace(*protocol)
	switch {
	case strings.EqualFold(*protocol, string(armnetwork.ProtocolTCP)):
		properties.Protocol = to.Ptr(armnetwork.ProbeProtocolTCP)
	case strings.EqualFold(*protocol, string(armnetwork.ProtocolHTTPS)):
		//HTTPS probe is only supported in standard loadbalancer
		//For backward compatibility,when unsupported protocol is used, fall back to tcp protocol in basic lb mode instead
		if !az.UseStandardLoadBalancer() {
			properties.Protocol = to.Ptr(armnetwork.ProbeProtocolTCP)
		} else {
			properties.Protocol = to.Ptr(armnetwork.ProbeProtocolHTTPS)
		}
	case strings.EqualFold(*protocol, string(armnetwork.ProtocolHTTP)):
		properties.Protocol = to.Ptr(armnetwork.ProbeProtocolHTTP)
	default:
		//For backward compatibility,when unsupported protocol is used, fall back to tcp protocol in basic lb mode instead
		properties.Protocol = to.Ptr(armnetwork.ProbeProtocolTCP)
	}

	// Select request path
	if strings.EqualFold(string(*properties.Protocol), string(armnetwork.ProtocolHTTPS)) || strings.EqualFold(string(*properties.Protocol), string(armnetwork.ProtocolHTTP)) {
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
			path = ptr.To(consts.HealthProbeDefaultRequestPath)
		}
		properties.RequestPath = path
	}

	properties.IntervalInSeconds, properties.ProbeThreshold, err = az.getHealthProbeConfigProbeIntervalAndNumOfProbe(serviceManifest, port.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to parse health probe config for port %d: %w", port.Port, err)
	}
	probe := &armnetwork.Probe{
		Name:       &lbrule,
		Properties: properties,
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
		probeInterval = ptr.To(consts.HealthProbeDefaultProbeInterval)
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
		numberOfProbes = ptr.To(consts.HealthProbeDefaultNumOfProbe)
	}
	return numberOfProbes, nil
}

func findProbe(probes []*armnetwork.Probe, probe *armnetwork.Probe) bool {
	for _, existingProbe := range probes {
		if strings.EqualFold(ptr.Deref(existingProbe.Name, ""), ptr.Deref(probe.Name, "")) &&
			ptr.Deref(existingProbe.Properties.Port, 0) == ptr.Deref(probe.Properties.Port, 0) &&
			strings.EqualFold(string(ptr.Deref(existingProbe.Properties.Protocol, "")), string(ptr.Deref(probe.Properties.Protocol, ""))) &&
			strings.EqualFold(ptr.Deref(existingProbe.Properties.RequestPath, ""), ptr.Deref(probe.Properties.RequestPath, "")) &&
			ptr.Deref(existingProbe.Properties.IntervalInSeconds, 0) == ptr.Deref(probe.Properties.IntervalInSeconds, 0) &&
			ptr.Deref(existingProbe.Properties.ProbeThreshold, 0) == ptr.Deref(probe.Properties.ProbeThreshold, 0) {
			return true
		}
	}
	return false
}

// keepSharedProbe ensures the shared probe will not be removed if there are more than 1 service referencing it.
func (az *Cloud) keepSharedProbe(
	service *v1.Service,
	lb armnetwork.LoadBalancer,
	expectedProbes []*armnetwork.Probe,
	wantLB bool,
) ([]*armnetwork.Probe, error) {
	logger := klog.Background().WithName("keepSharedProbe")
	var shouldConsiderRemoveSharedProbe bool
	if !wantLB {
		shouldConsiderRemoveSharedProbe = true
	}

	if lb.Properties != nil && lb.Properties.Probes != nil {
		for i, probe := range lb.Properties.Probes {
			if strings.EqualFold(ptr.Deref(probe.Name, ""), consts.SharedProbeName) {
				if !az.useSharedLoadBalancerHealthProbeMode() {
					shouldConsiderRemoveSharedProbe = true
				}
				if probe.Properties != nil && probe.Properties.LoadBalancingRules != nil {
					// Check if there's only one rule referencing this probe
					if len(probe.Properties.LoadBalancingRules) == 1 {
						ruleName, err := getLastSegment(*probe.Properties.LoadBalancingRules[0].ID, "/")
						if err != nil {
							logger.Error(err, "failed to parse load balancing rule name attached to health probe",
								"ruleName", *probe.Properties.LoadBalancingRules[0].ID, "healthProbe", *probe.ID)
						} else {
							// If the service owns the rule and is now a local service,
							// it means the service was switched from Cluster to Local
							if az.serviceOwnsRule(service, ruleName) && isLocalService(service) {
								logger.V(2).Info("service has switched from Cluster to Local, removing shared probe",
									"serviceName", getServiceName(service))
								// Remove the shared probe from the load balancer directly
								if lb.Properties != nil && lb.Properties.Probes != nil && i < len(lb.Properties.Probes) {
									lb.Properties.Probes = append(lb.Properties.Probes[:i], lb.Properties.Probes[i+1:]...)
								}
								return expectedProbes, nil
							}
						}
					}

					// Check other services referencing the probe
					for _, rule := range probe.Properties.LoadBalancingRules {
						ruleName, err := getLastSegment(*rule.ID, "/")
						if err != nil {
							logger.Error(err, "failed to parse load balancing rule name attached to health probe", "ruleName", *rule.ID, "healthProbe", *probe.ID)
							return []*armnetwork.Probe{}, err
						}
						if !az.serviceOwnsRule(service, ruleName) && shouldConsiderRemoveSharedProbe {
							logger.V(4).Info("there are load balancing rule of another service referencing the health probe, so the health probe should not be removed",
								"ruleID", *rule.ID, "probeID", *probe.ID)
							sharedProbe := az.buildClusterServiceSharedProbe()
							expectedProbes = append(expectedProbes, sharedProbe)
							return expectedProbes, nil
						}
					}
				}
			}
		}
	}
	return expectedProbes, nil
}
