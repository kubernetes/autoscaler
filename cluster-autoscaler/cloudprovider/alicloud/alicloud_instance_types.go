/*
Copyright 2016 The Kubernetes Authors.

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

package alicloud

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/golang/glog"
	"time"
)

type ecsInstance interface {
	DescribeInstanceTypes(req *ecs.DescribeInstanceTypesRequest) (*ecs.DescribeInstanceTypesResponse, error)
}

type instanceType struct {
	instanceTypeID string
	vcpu           int64
	memoryInBytes  int64
	gpu            int64
}

type instanceTypeModel struct {
	instanceType
	// TODO add price model .
}

// instanceWrapper will provide functions about
// instance types,price model and so on.
type instanceWrapper struct {
	ecsInstance
	InstanceTypeCache map[string]*instanceTypeModel
}

func (iw *instanceWrapper) getInstanceTypeById(typeId string) (*instanceType, error) {
	if instanceTypeModel := iw.FindInstanceType(typeId); instanceTypeModel != nil {
		return &instanceTypeModel.instanceType, nil
	}
	err := iw.RefreshCache()
	if err != nil {
		glog.Errorf("failed to refresh instance type cache,because of %s", err.Error())
		return nil, err
	}
	if instanceTypeModel := iw.FindInstanceType(typeId); instanceTypeModel != nil {
		return &instanceTypeModel.instanceType, nil
	}
	return nil, fmt.Errorf("failed to find the specific instance type by Id: %s", typeId)
}

func (iw *instanceWrapper) FindInstanceType(typeId string) *instanceTypeModel {
	if iw.InstanceTypeCache == nil || iw.InstanceTypeCache[typeId] == nil {
		return nil
	}
	return iw.InstanceTypeCache[typeId]
}

func (iw *instanceWrapper) RefreshCache() error {
	req := ecs.CreateDescribeInstanceTypesRequest()
	resp, err := iw.DescribeInstanceTypes(req)
	if err != nil {
		return err
	}
	if iw.InstanceTypeCache == nil {
		iw.InstanceTypeCache = make(map[string]*instanceTypeModel)
	}

	types := resp.InstanceTypes.InstanceType

	for _, item := range types {
		iw.InstanceTypeCache[item.InstanceTypeId] = &instanceTypeModel{
			instanceType{
				instanceTypeID: item.InstanceTypeId,
				vcpu:           int64(item.CpuCoreCount),
				memoryInBytes:  int64(item.MemorySize * 1024 * 1024 * 1024),
				gpu:            int64(item.GPUAmount),
			},
		}
	}
	return nil
}

func NewInstanceWrapper(cfg *CloudConfig) (*instanceWrapper, error) {
	if cfg.IsValid() == false {
		return nil, fmt.Errorf("your cloud config is not valid")
	}
	iw := &instanceWrapper{}
	if cfg.STSEnabled == true {
		go func(iw *instanceWrapper, cfg *CloudConfig) {
			timer := time.NewTicker(REFRESH_CLIENT_INTERVAL)
			defer timer.Stop()
			for {
				select {
				case <-timer.C:
					client, err := GetEcsClient(cfg)
					if err == nil {
						iw.ecsInstance = client
					}
				}
			}
		}(iw, cfg)
	}
	client, err := GetEcsClient(cfg)
	if err == nil {
		iw.ecsInstance = client
	}
	return iw, err
}

func GetEcsClient(cfg *CloudConfig) (client *ecs.Client, err error) {
	region := cfg.GetRegion()
	if cfg.STSEnabled == true {
		auth, err := cfg.GetSTSToken()
		if err != nil {
			glog.Errorf("failed to get sts token from metadata,because of %s", err.Error())
			return nil, err
		}
		client, err = ecs.NewClientWithStsToken(region, auth.AccessKeyId, auth.AccessKeySecret, auth.SecurityToken)
		if err != nil {
			glog.Errorf("failed to create client with sts in metadata,because of %s", err.Error())
		}
	} else {
		client, err = ecs.NewClientWithAccessKey(region, cfg.AccessKeyID, cfg.AccessKeySecret)
		if err != nil {
			glog.Errorf("failed to create ecs client with AccessKeyId and AccessKeySecret,because of %s", err.Error())
		}
	}
	return
}
