package gce

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gce "google.golang.org/api/compute/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func TestMachineCache(t *testing.T) {
	type CacheQuery struct {
		machineType string
		zone        string
		machine     *gce.MachineType
		err         error
	}
	testCases := []struct {
		desc     string
		machines []CacheQuery
		want     map[MachineTypeKey]uint64
		wantErr  []MachineTypeKey
	}{
		{
			desc: "replacement",
			machines: []CacheQuery{
				{
					machineType: "e2-standard-2",
					zone:        "myzone",
					machine:     &gce.MachineType{Id: 1},
				},
				{
					machineType: "e2-standard-2",
					zone:        "myzone",
					machine:     &gce.MachineType{Id: 1},
				},
				{
					machineType: "e2-standard-2",
					zone:        "myzone",
					machine:     &gce.MachineType{Id: 2},
				},
				{
					machineType: "e2-standard-2",
					zone:        "myzone",
					machine:     &gce.MachineType{Id: 2},
				},
				{
					machineType: "e2-standard-4",
					zone:        "myzone2",
					err:         errors.New("error"),
				},
			},
			want: map[MachineTypeKey]uint64{
				MachineTypeKey{
					MachineType: "e2-standard-2",
					Zone:        "myzone",
				}: 2,
			},
			wantErr: []MachineTypeKey{
				{
					Zone:        "myzone2",
					MachineType: "e2-standard-4",
				},
			},
		},
	}
	c := NewGceCache(nil)
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			for _, m := range tc.machines {
				if m.err != nil {
					c.AddMachineToCacheWithError(m.machineType, m.zone, m.err)
					continue
				}
				c.AddMachineToCache(m.machineType, m.zone, m.machine)
			}
			for mt, wantId := range tc.want {
				m, err := c.GetMachineFromCache(mt.MachineType, mt.Zone)
				if err != nil {
					t.Errorf("Did not expect error for machine type = %q, zone = %q", mt.MachineType, mt.Zone)
				}
				if m.Id != wantId {
					t.Errorf("Wanted id %d but got id %d for machine type = %q, zone = %q", wantId, m.Id, mt.MachineType, mt.Zone)
				}
			}
			for _, mt := range tc.wantErr {
				_, err := c.GetMachineFromCache(mt.MachineType, mt.Zone)
				if err == nil {
					t.Errorf("Wanted an error but got no error for machine type = %q, zone = %q", mt.MachineType, mt.Zone)
				}
			}
		})
	}
	gceManagerMock := &gceManagerMock{}
	gce := &GceCloudProvider{
		gceManager: gceManagerMock,
	}
	mig := &gceMig{gceRef: GceRef{Name: "ng1"}}
	gceManagerMock.On("GetMigs").Return([]Mig{mig}).Once()
	result := gce.NodeGroups()
	assert.Equal(t, []cloudprovider.NodeGroup{mig}, result)
	mock.AssertExpectationsForObjects(t, gceManagerMock)
}
