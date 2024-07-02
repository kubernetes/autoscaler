/*
Copyright 2021 The Kubernetes Authors.

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

// Package consts stages all the consts under pkg/.
package consts

import (
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// IsK8sServiceHasHAModeEnabled return if HA Mode is enabled in kubernetes service annotations
func IsK8sServiceHasHAModeEnabled(service *v1.Service) bool {
	return expectAttributeInSvcAnnotationBeEqualTo(service.Annotations, ServiceAnnotationLoadBalancerEnableHighAvailabilityPorts, TrueAnnotationValue)
}

// IsK8sServiceUsingInternalLoadBalancer return if service is using an internal load balancer.
func IsK8sServiceUsingInternalLoadBalancer(service *v1.Service) bool {
	return expectAttributeInSvcAnnotationBeEqualTo(service.Annotations, ServiceAnnotationLoadBalancerInternal, TrueAnnotationValue)
}

// IsK8sServiceDisableLoadBalancerFloatingIP return if floating IP in load balancer is disabled in kubernetes service annotations
func IsK8sServiceDisableLoadBalancerFloatingIP(service *v1.Service) bool {
	return expectAttributeInSvcAnnotationBeEqualTo(service.Annotations, ServiceAnnotationDisableLoadBalancerFloatingIP, TrueAnnotationValue)
}

// GetHealthProbeConfigOfPortFromK8sSvcAnnotation get health probe configuration for port
func GetHealthProbeConfigOfPortFromK8sSvcAnnotation(annotations map[string]string, port int32, key HealthProbeParams, validators ...BusinessValidator) (*string, error) {
	return GetAttributeValueInSvcAnnotation(annotations, BuildHealthProbeAnnotationKeyForPort(port, key), validators...)
}

// IsHealthProbeRuleOnK8sServicePortDisabled return if port is for health probe only
func IsHealthProbeRuleOnK8sServicePortDisabled(annotations map[string]string, port int32) (bool, error) {
	return expectAttributeInSvcAnnotationBeEqualTo(annotations, BuildAnnotationKeyForPort(port, PortAnnotationNoHealthProbeRule), TrueAnnotationValue), nil
}

// IsHealthProbeRuleOnK8sServicePortDisabled return if port is for health probe only
func IsLBRuleOnK8sServicePortDisabled(annotations map[string]string, port int32) (bool, error) {
	return expectAttributeInSvcAnnotationBeEqualTo(annotations, BuildAnnotationKeyForPort(port, PortAnnotationNoLBRule), TrueAnnotationValue), nil
}

// IsPLSProxyProtocolEnabled return true if ServiceAnnotationPLSProxyProtocol is true
func IsPLSProxyProtocolEnabled(annotations map[string]string) bool {
	return expectAttributeInSvcAnnotationBeEqualTo(annotations, ServiceAnnotationPLSProxyProtocol, TrueAnnotationValue)
}

// IsPLSEnabled return true if ServiceAnnotationPLSCreation is true
func IsPLSEnabled(annotations map[string]string) bool {
	return expectAttributeInSvcAnnotationBeEqualTo(annotations, ServiceAnnotationPLSCreation, TrueAnnotationValue)
}

// IsTCPResetDisabled return true if ServiceAnnotationDisableTCPReset is true
func IsTCPResetDisabled(annotations map[string]string) bool {
	return expectAttributeInSvcAnnotationBeEqualTo(annotations, ServiceAnnotationDisableTCPReset, TrueAnnotationValue)
}

// Getint32ValueFromK8sSvcAnnotation get health probe configuration for port
func Getint32ValueFromK8sSvcAnnotation(annotations map[string]string, key string, validators ...Int32BusinessValidator) (*int32, error) {
	val, err := GetAttributeValueInSvcAnnotation(annotations, key)
	if err == nil && val != nil {
		return extractInt32FromString(*val, validators...)
	}
	return nil, err
}

// BuildAnnotationKeyForPort get health probe configuration key for port
func BuildAnnotationKeyForPort(port int32, key PortParams) string {
	return fmt.Sprintf(PortAnnotationPrefixPattern, port, string(key))
}

// BuildHealthProbeAnnotationKeyForPort get health probe configuration key for port
func BuildHealthProbeAnnotationKeyForPort(port int32, key HealthProbeParams) string {
	return BuildAnnotationKeyForPort(port, PortParams(fmt.Sprintf(HealthProbeAnnotationPrefixPattern, key)))
}

// GetInt32HealthProbeConfigOfPortFromK8sSvcAnnotation get health probe configuration for port
func GetInt32HealthProbeConfigOfPortFromK8sSvcAnnotation(annotations map[string]string, port int32, key HealthProbeParams, validators ...Int32BusinessValidator) (*int32, error) {
	return Getint32ValueFromK8sSvcAnnotation(annotations, BuildHealthProbeAnnotationKeyForPort(port, key), validators...)
}

// Int32BusinessValidator is validator function which is invoked after values are parsed in order to make sure input value meets the businees need.
type Int32BusinessValidator func(*int32) error

// getInt32FromAnnotations parse integer value from annotation and return an reference to int32 object
func extractInt32FromString(val string, businessValidator ...Int32BusinessValidator) (*int32, error) {
	val = strings.TrimSpace(val)
	errKey := fmt.Sprintf("%s value must be a whole number", val)
	toInt, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("error value: %w: %s", err, errKey)
	}
	parsedInt := int32(toInt)
	for _, validator := range businessValidator {
		if validator != nil {
			err := validator(&parsedInt)
			if err != nil {
				return nil, fmt.Errorf("error parsing value: %w", err)
			}
		}
	}
	return &parsedInt, nil
}

// BusinessValidator is validator function which is invoked after values are parsed in order to make sure input value meets the businees need.
type BusinessValidator func(*string) error

// GetAttributeValueInSvcAnnotation get value in annotation map using key
func GetAttributeValueInSvcAnnotation(annotations map[string]string, key string, validators ...BusinessValidator) (*string, error) {
	if l, found := annotations[key]; found {
		for _, validateFunc := range validators {
			if validateFunc != nil {
				if err := validateFunc(&l); err != nil {
					return nil, err
				}
			}
		}
		return &l, nil
	}
	return nil, nil
}

// expectAttributeInSvcAnnotation get key in svc annotation and compare with target value
func expectAttributeInSvcAnnotationBeEqualTo(annotations map[string]string, key string, value string) bool {
	if l, err := GetAttributeValueInSvcAnnotation(annotations, key); err == nil && l != nil {
		return strings.EqualFold(*l, value)
	}
	return false
}

// getLoadBalancerConfigurationsNames parse the annotation and return the names of the load balancer configurations.
func GetLoadBalancerConfigurationsNames(service *v1.Service) []string {
	var names []string
	for key, lbConfig := range service.Annotations {
		if strings.EqualFold(key, ServiceAnnotationLoadBalancerConfigurations) {
			names = append(names, strings.Split(lbConfig, ",")...)
		}
	}
	for i := range names {
		names[i] = strings.ToLower(strings.TrimSpace(names[i]))
	}
	return names
}
