/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/core"
	"k8s.io/klog/v2"
)

// ShapeGetter returns the oci shape attributes for the pool.
type ShapeGetter interface {
	GetNodePoolShape(*oke.NodePool, int64) (*Shape, error)
	GetInstancePoolShape(pool *core.InstancePool) (*Shape, error)
	Refresh()
}

// ShapeClient is an interface around the GetInstanceConfiguration and ListShapes calls.
type ShapeClient interface {
	GetInstanceConfiguration(context.Context, core.GetInstanceConfigurationRequest) (core.GetInstanceConfigurationResponse, error)
	ListShapes(context.Context, core.ListShapesRequest) (core.ListShapesResponse, error)
}

// ShapeClientImpl is the implementation for fetching shape information.
type ShapeClientImpl struct {
	// Can fetch instance configs (flexible shapes)
	ComputeMgmtClient core.ComputeManagementClient
	// Can fetch shapes directly
	ComputeClient core.ComputeClient
}

// GetInstanceConfiguration gets the instance configuration.
func (cc ShapeClientImpl) GetInstanceConfiguration(ctx context.Context, req core.GetInstanceConfigurationRequest) (core.GetInstanceConfigurationResponse, error) {
	return cc.ComputeMgmtClient.GetInstanceConfiguration(ctx, req)
}

// ListShapes lists the shapes.
func (cc ShapeClientImpl) ListShapes(ctx context.Context, req core.ListShapesRequest) (core.ListShapesResponse, error) {
	return cc.ComputeClient.ListShapes(ctx, req)
}

// Shape includes the resource attributes of a given shape which should be used
// for constructing node templates.
type Shape struct {
	Name                    string
	CPU                     float32
	GPU                     int
	MemoryInBytes           float32
	EphemeralStorageInBytes float32
}

// CreateShapeGetter creates a new oci shape getter.
func CreateShapeGetter(shapeClient ShapeClient) ShapeGetter {
	return &shapeGetterImpl{
		shapeClient: shapeClient,
		cache:       map[string]*Shape{},
	}
}

type shapeGetterImpl struct {
	shapeClient ShapeClient
	cache       map[string]*Shape
	mu          sync.Mutex
}

// Refresh clears out the cache to be populated again as the pool shapes are re-requested
func (osf *shapeGetterImpl) Refresh() {
	// For now, just clear the cache
	osf.cache = map[string]*Shape{}
}

// GetNodePoolShape gets the shape by querying the node pool's configuration
func (osf *shapeGetterImpl) GetNodePoolShape(np *oke.NodePool, ephemeralStorage int64) (*Shape, error) {
	shapeName := *np.NodeShape
	if np.NodeShapeConfig != nil {
		return &Shape{
			CPU: *np.NodeShapeConfig.Ocpus * 2,
			// num_bytes * kilo * mega * giga
			MemoryInBytes:           *np.NodeShapeConfig.MemoryInGBs * 1024 * 1024 * 1024,
			GPU:                     0,
			EphemeralStorageInBytes: float32(ephemeralStorage),
		}, nil
	}

	osf.mu.Lock()
	defer osf.mu.Unlock()

	// check cache first
	shape, ok := osf.cache[shapeName]
	if ok {
		return shape, nil
	}

	// refresh cache if we have a miss.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := osf.shapeClient.ListShapes(ctx, core.ListShapesRequest{
		CompartmentId: np.CompartmentId,
		Limit:         common.Int(500),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to ListShapes")
	}

	// Update the cache based on latest results
	for _, s := range resp.Items {
		osf.cache[*s.Shape] = &Shape{
			CPU:                     getFloat32(s.Ocpus) * 2, // convert ocpu to vcpu
			GPU:                     getInt(s.Gpus),
			MemoryInBytes:           getFloat32(s.MemoryInGBs) * 1024 * 1024 * 1024,
			EphemeralStorageInBytes: float32(ephemeralStorage),
		}
	}

	// fetch value from updated cache... if it exists.
	shape, ok = osf.cache[shapeName]
	if ok {
		return shape, nil
	}

	return nil, fmt.Errorf("shape %q does not exist", shapeName)
}

// GetInstancePoolShape gets the shape by querying the instance pool's configuration
func (osf *shapeGetterImpl) GetInstancePoolShape(ip *core.InstancePool) (*Shape, error) {

	// First, check instance pool shape cache
	shape, ok := osf.cache[*ip.Id]
	if ok {
		return shape, nil
	}

	klog.V(5).Info("fetching shape configuration details for instance-pool " + *ip.Id)
	shape = &Shape{}

	instanceConfig, err := osf.shapeClient.GetInstanceConfiguration(context.Background(), core.GetInstanceConfigurationRequest{
		InstanceConfigurationId: ip.InstanceConfigurationId,
	})
	if err != nil {
		return nil, err
	}

	if instanceConfig.InstanceDetails == nil {
		return nil, fmt.Errorf("instance configuration details for instance %s has not been set", *ip.Id)
	}

	if instanceDetails, ok := instanceConfig.InstanceDetails.(core.ComputeInstanceDetails); ok {
		// flexible shape use details or look up the static shape details below.
		if instanceDetails.LaunchDetails != nil && instanceDetails.LaunchDetails.ShapeConfig != nil {
			if instanceDetails.LaunchDetails.Shape != nil {
				shape.Name = *instanceDetails.LaunchDetails.Shape
			}
			if instanceDetails.LaunchDetails.ShapeConfig.Ocpus != nil {
				shape.CPU = *instanceDetails.LaunchDetails.ShapeConfig.Ocpus
				// Minimum amount of memory unless explicitly set higher
				shape.MemoryInBytes = *instanceDetails.LaunchDetails.ShapeConfig.Ocpus * 1024 * 1024 * 1024
			}
			if instanceDetails.LaunchDetails.ShapeConfig.MemoryInGBs != nil {
				shape.MemoryInBytes = *instanceDetails.LaunchDetails.ShapeConfig.MemoryInGBs * 1024 * 1024 * 1024
			}
		} else {
			// Fetch the shape object by name
			var page *string
			var everyShape []core.Shape
			for {
				// List all available shapes
				lisShapesReq := core.ListShapesRequest{}
				lisShapesReq.CompartmentId = instanceConfig.CompartmentId
				lisShapesReq.Page = page
				lisShapesReq.Limit = common.Int(50)

				listShapes, err := osf.shapeClient.ListShapes(context.Background(), lisShapesReq)
				if err != nil {
					return nil, err
				}

				everyShape = append(everyShape, listShapes.Items...)

				if page = listShapes.OpcNextPage; listShapes.OpcNextPage == nil {
					break
				}
			}

			for _, nextShape := range everyShape {
				if *nextShape.Shape == *instanceDetails.LaunchDetails.Shape {
					shape.Name = *nextShape.Shape
					if nextShape.Ocpus != nil {
						shape.CPU = *nextShape.Ocpus
					}
					if nextShape.MemoryInGBs != nil {
						shape.MemoryInBytes = *nextShape.MemoryInGBs * 1024 * 1024 * 1024
					}
					if nextShape.Gpus != nil {
						shape.GPU = *nextShape.Gpus
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("(compute) instance configuration for instance-pool %s not found", *ip.Id)
	}

	// Didn't find a match
	if shape.Name == "" {
		return nil, fmt.Errorf("shape information for instance-pool %s not found", *ip.Id)
	}

	osf.cache[*ip.Id] = shape
	return shape, nil
}

// getFloat32 is a helper to get a float32 pointer value or default to 0.
func getFloat32(f *float32) float32 {
	if f == nil {
		return 0.0
	}
	return *f
}

// getInt is a helper to get an int pointer value or default to 0.
func getInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
