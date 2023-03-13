package wrapper

import (
	"reflect"
	"testing"
)

func TestMaxResourceLimitReached(t *testing.T) {
	tests := []struct {
		name        string
		resources   []string
		wantReasons []string
	}{
		{
			name:        "simple test",
			resources:   []string{"gpu"},
			wantReasons: []string{"max cluster gpu limit reached"},
		},
		{
			name:        "multiple resources",
			resources:   []string{"gpu1", "gpu3", "tpu", "ram"},
			wantReasons: []string{"max cluster gpu1, gpu3, tpu, ram limit reached"},
		},
		{
			name:        "no resources",
			wantReasons: []string{"max cluster  limit reached"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxResourceLimitReached(tt.resources); !reflect.DeepEqual(got.Reasons(), tt.wantReasons) {
				t.Errorf("MaxResourceLimitReached(%v) = %v, want %v", tt.resources, got.Reasons(), tt.wantReasons)
			}
		})
	}
}
