package oci

import (
	"testing"
)

func TestSetProviderID(t *testing.T) {
	err := setNodeProviderID(nil, "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}
