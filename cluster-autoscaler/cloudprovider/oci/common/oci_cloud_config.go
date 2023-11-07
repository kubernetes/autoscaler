/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/gcfg.v1"
	ipconsts "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/instancepools/consts"
	npconsts "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/nodepools/consts"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"k8s.io/klog/v2"
)

// CloudConfig holds the cloud config for OCI provider.
type CloudConfig struct {
	Global struct {
		RefreshInterval        time.Duration `gcfg:"refresh-interval"`
		CompartmentID          string        `gcfg:"compartment-id"`
		Region                 string        `gcfg:"region"`
		UseInstancePrinciples  bool          `gcfg:"use-instance-principals"`
		UseNonMemberAnnotation bool          `gcfg:"use-non-member-annotation"`
	}
}

// CreateCloudConfig creates a CloudConfig object based on a file or env vars
func CreateCloudConfig(cloudConfigPath string, configProvider common.ConfigurationProvider, implType string) (*CloudConfig, error) {
	var cloudConfig = &CloudConfig{}

	// cloudConfigPath is the optional file of variables passed in with the --cloud-config flag, which takes precedence over environment variables
	if cloudConfigPath != "" {
		config, fileErr := os.Open(cloudConfigPath)
		if fileErr != nil {
			klog.Fatalf("could not open cloud provider configuration %s: %#v", cloudConfigPath, fileErr)
		}
		defer config.Close()
		if config != nil {
			if err := gcfg.ReadInto(cloudConfig, config); err != nil {
				klog.Errorf("could not read config: %v", err)
				return nil, err
			}
		}
	}
	// Fall back to environment variables
	if cloudConfig.Global.CompartmentID == "" {
		cloudConfig.Global.CompartmentID = os.Getenv(ipconsts.OciCompartmentEnvVar)
	} else if !cloudConfig.Global.UseInstancePrinciples {
		if os.Getenv(ipconsts.OciUseInstancePrincipalEnvVar) == "true" {
			cloudConfig.Global.UseInstancePrinciples = true
		}
		if os.Getenv(ipconsts.OciRegionEnvVar) != "" {
			cloudConfig.Global.Region = os.Getenv(ipconsts.OciRegionEnvVar)
		}
	}
	if cloudConfig.Global.RefreshInterval == 0 {
		if os.Getenv(ipconsts.OciRefreshInterval) != "" {
			klog.V(4).Infof("using a custom cache refresh interval %v...", os.Getenv(ipconsts.OciRefreshInterval))
			cloudConfig.Global.RefreshInterval, _ = time.ParseDuration(os.Getenv(ipconsts.OciRefreshInterval))
		} else {
			if implType == npconsts.OciNodePoolResourceIdent {
				cloudConfig.Global.RefreshInterval = npconsts.DefaultRefreshInterval
			} else {
				cloudConfig.Global.RefreshInterval = ipconsts.DefaultRefreshInterval
			}
		}
	}

	// this env var is only relevant for instance pools
	if os.Getenv(ipconsts.OciUseNonPoolMemberAnnotationEnvVar) == "true" {
		cloudConfig.Global.UseNonMemberAnnotation = true
	}

	cloudConfig.Global.CompartmentID = os.Getenv(ipconsts.OciCompartmentEnvVar)

	// Not passed by --cloud-config or environment variable, attempt to use the tenancy ID as the compartment ID
	if cloudConfig.Global.CompartmentID == "" {
		tenancyID, err := configProvider.TenancyOCID()
		if err != nil {
			return nil, errors.Wrap(err, "unable to retrieve tenancy ID")
		}
		cloudConfig.Global.CompartmentID = tenancyID
	}
	return cloudConfig, nil
}
