/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"testing"
)

func TestSetProviderID(t *testing.T) {
	err := SetNodeProviderID(nil, "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}
