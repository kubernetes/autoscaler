// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package mcm is used to provide the core functionalities of machine-controller-manager
package mcm

import (
	clientset "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned"
	restclient "k8s.io/client-go/rest"

	"k8s.io/klog/v2"
)

// MachineClientBuilder allows you to get clients and configs for machine controllers
type MachineClientBuilder interface {
	// Config returns a new restclient.Config with the given user agent name.
	Config(name string) (*restclient.Config, error)
	// ConfigOrDie return a new restclient.Config with the given user agent
	// name, or logs a fatal error.
	ConfigOrDie(name string) *restclient.Config
	// Client returns a new clientset.Interface with the given user agent
	// name.
	Client(name string) (clientset.Interface, error)
	// ClientOrDie returns a new clientset.Interface with the given user agent
	// name or logs a fatal error, destroying the computer and killing the
	// operator and programmer.
	ClientOrDie(name string) clientset.Interface
}

// MachineControllerClientBuilder returns a fixed client with different user agents
type MachineControllerClientBuilder struct {
	// ClientConfig is a skeleton config to clone and use as the basis for each controller client
	ClientConfig *restclient.Config
}

// Config returns a new restclient.Config with the given user agent name.
func (b MachineControllerClientBuilder) Config(name string) (*restclient.Config, error) {
	clientConfig := *b.ClientConfig
	return restclient.AddUserAgent(&clientConfig, name), nil
}

// ConfigOrDie return a new restclient.Config with the given user agent
// name, or logs a fatal error.
func (b MachineControllerClientBuilder) ConfigOrDie(name string) *restclient.Config {
	clientConfig, err := b.Config(name)
	if err != nil {
		klog.Fatal(err)
	}
	return clientConfig
}

// Client returns a new clientset.Interface with the given user agent
// name.
func (b MachineControllerClientBuilder) Client(name string) (clientset.Interface, error) {
	clientConfig, err := b.Config(name)
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(clientConfig)
}

// ClientOrDie returns a new clientset.Interface with the given user agent
// name or logs a fatal error, destroying the computer and killing the
// operator and programmer.
func (b MachineControllerClientBuilder) ClientOrDie(name string) clientset.Interface {
	client, err := b.Client(name)
	if err != nil {
		klog.Fatal(err)
	}
	return client
}
