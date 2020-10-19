/*
Copyright 2020 The Kubernetes Authors.

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

package cloudstack

import (
	"fmt"
	"os"

	"gopkg.in/gcfg.v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/cloudstack/service"
	klog "k8s.io/klog/v2"
)

type managerConfig struct {
	asg       *asg
	acsConfig *service.Config
	service   service.CKSService
	file      string
}

type option func(*managerConfig)

func withConfigFile(file string) option {
	return func(config *managerConfig) {
		config.file = file
	}
}

func withASG(asg *asg) option {
	return func(config *managerConfig) {
		config.asg = asg
	}
}

func withCKSService(service service.CKSService) option {
	return func(config *managerConfig) {
		config.service = service
	}
}

func readConfig(file string) (*CSConfig, error) {

	cfg := &CSConfig{}

	fp, err := os.Open(file)
	if err != nil || fp == nil {
		klog.Fatalf("Unable to open cloud provider configuration (cloud-config) %s : %v", file, err)
	}

	defer func() {
		err = fp.Close()
		if err != nil {
			klog.Warningf("Unable to close config : %v", err)
		}
	}()

	if err := gcfg.ReadInto(cfg, fp); err != nil {
		return nil, fmt.Errorf("Unable to parse cloud provider config : %v", err)
	}

	if cfg.Global.APIURL == "" || cfg.Global.APIKey == "" || cfg.Global.SecretKey == "" {
		return nil, fmt.Errorf("Missing api-url, api-key or secret-key in the cloud-config")
	}
	return cfg, nil
}

func createConfig(opts ...option) (*managerConfig, error) {
	cfg := &managerConfig{}
	for _, fn := range opts {
		fn(cfg)
	}

	if cfg.asg == nil {
		cfg.asg = &asg{}
	}

	if cfg.file != "" {
		config, err := readConfig(cfg.file)
		if err != nil {
			return nil, err
		}
		cfg.acsConfig = &service.Config{
			APIKey:    config.Global.APIKey,
			SecretKey: config.Global.SecretKey,
			Endpoint:  config.Global.APIURL,
		}
	}

	if cfg.service == nil {
		cfg.service = service.NewCKSService(cfg.acsConfig)
	}
	return cfg, nil
}
