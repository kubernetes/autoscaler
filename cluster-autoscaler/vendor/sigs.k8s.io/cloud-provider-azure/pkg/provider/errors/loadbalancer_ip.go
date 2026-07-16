/*
Copyright 2026 The Kubernetes Authors.

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

package providererrors

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidLoadBalancerIP identifies a LoadBalancer IP that is not a valid IP address.
	ErrInvalidLoadBalancerIP = errors.New("invalid LoadBalancer IP")
	// ErrNonPublicLoadBalancerIP identifies a LoadBalancer IP that cannot be used as an Azure Public IP.
	ErrNonPublicLoadBalancerIP = errors.New("non-public LoadBalancer IP")
)

type serviceLoadBalancerIPError struct {
	msg string
	err error
}

func (e *serviceLoadBalancerIPError) Error() string {
	return e.msg
}

func (e *serviceLoadBalancerIPError) Unwrap() error {
	return e.err
}

// NewInvalidLoadBalancerIPError returns an error for a syntactically invalid LoadBalancer IP.
func NewInvalidLoadBalancerIPError(ip string) error {
	return fmt.Errorf("%w %q", ErrInvalidLoadBalancerIP, ip)
}

// NewNonPublicLoadBalancerIPError returns an error for a LoadBalancer IP that is not a public IP candidate.
func NewNonPublicLoadBalancerIPError(ip string) error {
	return fmt.Errorf("%w %q", ErrNonPublicLoadBalancerIP, ip)
}

// IsLoadBalancerIPValidationError reports whether err is a known LoadBalancer IP validation error.
func IsLoadBalancerIPValidationError(err error) bool {
	return errors.Is(err, ErrInvalidLoadBalancerIP) || errors.Is(err, ErrNonPublicLoadBalancerIP)
}

// NewExternalServiceLoadBalancerIPError returns a clean external Service error while preserving the validation cause.
func NewExternalServiceLoadBalancerIPError(serviceName, loadBalancerIP string, err error) error {
	if errors.Is(err, ErrInvalidLoadBalancerIP) {
		return &serviceLoadBalancerIPError{
			msg: fmt.Sprintf("external Service %q has invalid LoadBalancer IP %q", serviceName, loadBalancerIP),
			err: err,
		}
	}
	if errors.Is(err, ErrNonPublicLoadBalancerIP) {
		return &serviceLoadBalancerIPError{
			msg: fmt.Sprintf("external Service %q cannot use non-public LoadBalancer IP %q; use an internal LoadBalancer annotation or a valid Azure Public IP address", serviceName, loadBalancerIP),
			err: err,
		}
	}
	return err
}
