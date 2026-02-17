package duplocloud

import (
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
)

func TestRdsInstanceToState_StorageAutoscaling(t *testing.T) {
	tests := []struct {
		name                    string
		isEnabled               bool
		maxAllocatedStorage     int
		allocatedStorage        int
		wantEnable              bool
		wantMaxAllocatedStorage int
	}{
		{
			name:                    "enabled",
			isEnabled:               true,
			maxAllocatedStorage:     100,
			allocatedStorage:        20,
			wantEnable:              true,
			wantMaxAllocatedStorage: 100,
		},
		{
			name:                    "disabled",
			isEnabled:               false,
			maxAllocatedStorage:     20,
			allocatedStorage:        20,
			wantEnable:              false,
			wantMaxAllocatedStorage: 20,
		},
		{
			name:                    "disabled_zero_max",
			isEnabled:               false,
			maxAllocatedStorage:     0,
			allocatedStorage:        20,
			wantEnable:              false,
			wantMaxAllocatedStorage: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := resourceDuploRdsInstance().TestResourceData()
			duploObject := &duplosdk.DuploRdsInstance{
				IsAutoScalingEnabled: tt.isEnabled,
				MaxAllocatedStorage:  tt.maxAllocatedStorage,
				AllocatedStorage:     tt.allocatedStorage,
			}

			jo := rdsInstanceToState(duploObject, d)

			sa, ok := jo["storage_autoscaling"]
			if !ok {
				t.Fatal("expected storage_autoscaling to be present in state")
			}
			saList, ok := sa.([]interface{})
			if !ok {
				t.Fatal("expected storage_autoscaling to be []interface{}")
			}
			if len(saList) != 1 {
				t.Fatalf("expected storage_autoscaling to have 1 element, got %d", len(saList))
			}
			saMap, ok := saList[0].(map[string]interface{})
			if !ok {
				t.Fatal("expected storage_autoscaling element to be map[string]interface{}")
			}
			if saMap["enable"] != tt.wantEnable {
				t.Errorf("expected storage_autoscaling.enable = %v, got %v", tt.wantEnable, saMap["enable"])
			}
			if saMap["max_allocated_storage"] != tt.wantMaxAllocatedStorage {
				t.Errorf("expected storage_autoscaling.max_allocated_storage = %v, got %v", tt.wantMaxAllocatedStorage, saMap["max_allocated_storage"])
			}
		})
	}
}

func TestRdsInstanceToState_NilObject(t *testing.T) {
	jo := rdsInstanceToState(nil, nil)
	if jo != nil {
		t.Errorf("expected nil result for nil input, got %v", jo)
	}
}
