// Copyright (c) 2016, 2018, 2022, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"crypto/tls"
	"errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v55/common"
	"net/http"
	"testing"
)

var customTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		ServerName: "test",
	},
}

func TestNewDispatcherModifier_NoInitialModifier(t *testing.T) {
	modifier := newDispatcherModifier(nil)
	initialClient := &http.Client{}
	returnClient, err := modifier.Modify(initialClient)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(initialClient, returnClient)
}

func TestNewDispatcherModifier_InitialModifier(t *testing.T) {
	modifier := newDispatcherModifier(setCustomCAPool)
	initialClient := &http.Client{}
	returnDispatcher, err := modifier.Modify(initialClient)
	assert.Nil(t, err)
	returnClient := returnDispatcher.(*http.Client)
	assert.ObjectsAreEqual(returnClient.Transport, customTransport)
}

func TestNewDispatcherModifier_ModifierFails(t *testing.T) {
	modifier := newDispatcherModifier(modifierGoneWrong)
	initialClient := &http.Client{}
	returnClient, err := modifier.Modify(initialClient)
	assert.NotNil(t, err)
	assert.Nil(t, returnClient)
}

func setCustomCAPool(dispatcher common.HTTPRequestDispatcher) (common.HTTPRequestDispatcher, error) {
	client := dispatcher.(*http.Client)
	client.Transport = customTransport
	return client, nil
}

func modifierGoneWrong(dispatcher common.HTTPRequestDispatcher) (common.HTTPRequestDispatcher, error) {
	return nil, errors.New("uh oh")
}
