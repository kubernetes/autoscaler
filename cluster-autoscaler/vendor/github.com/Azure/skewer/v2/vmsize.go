package skewer

import (
	"fmt"
	"regexp"
	"strconv"
)

// This file adds support for more capabilities based on VM naming conventions that includes vmsize parsing.
// VM naming conventions are documented at: https://docs.microsoft.com/en-us/azure/virtual-machines/vm-naming-conventions
// Note: Some common capabilities like familyName and VCPUs, which can also be
// fetched using the ResourceSKU API, are not included here. They can be found in sku.go.

var skuSizeScheme = regexp.MustCompile(
	`^([A-Z])([A-Z]?)([A-Z]?)([0-9]+)-?((?:[0-9]+)?)((?:[abcdeilmtspPr]+|C+|NP)?)_?(?:([A-Z][0-9]+)_?)?(_cc_)?(_[0-9]+_)?(_MI300X_)?(_H100_)?((?:[vV][1-9])?)?(_Promo)?$`,
)

// unParsableVMSizes map holds vmSize strings that cannot be easily parsed with skuSizeScheme.
var unParsableVMSizes = map[string]VMSizeType{
	"M416s_8_v2": {
		Family:                      "M",
		Subfamily:                   nil,
		Cpus:                        "416",
		CpusConstrained:             nil,
		AdditiveFeatures:            []rune{'s'},
		AcceleratorType:             nil,
		ConfidentialChildCapability: false,
		Version:                     "v2",
		PromoVersion:                false,
		Series:                      "Ms_v2",
	},
}

type VMSizeType struct {
	Family                      string
	Subfamily                   *string
	Cpus                        string
	CpusConstrained             *string
	AdditiveFeatures            []rune
	AcceleratorType             *string
	ConfidentialChildCapability bool
	Version                     string
	PromoVersion                bool
	MI300Series                 bool
	H100Series                  bool
	Series                      string
}

// parseVMSize parses the VM size and returns the parts as a map.
func parseVMSize(vmSizeName string) ([]string, error) {
	parts := skuSizeScheme.FindStringSubmatch(vmSizeName)
	if len(parts) < 10 {
		return nil, fmt.Errorf("could not parse VM size %s", vmSizeName)
	}
	return parts, nil
}

// GetVMSize is a helper function used by GetVMSize() in sku.go
func GetVMSize(vmSizeName string) (*VMSizeType, error) {
	vmSize := VMSizeType{}

	parts, err := parseVMSize(vmSizeName)
	if err != nil {
		if vmSizeVal, ok := unParsableVMSizes[vmSizeName]; ok {
			return &vmSizeVal, nil
		}
		return nil, err
	}

	// [Family] - ([A-Z]): Captures a single uppercase letter.
	vmSize.Family = parts[1]

	// [Sub-family]* - ([A-Z]?): Optionally captures another uppercase letter.
	if len(parts[2]) > 0 {
		var subfamilyStr string
		if len(parts[3]) > 0 {
			subfamilyStr = parts[2] + parts[3]
		} else {
			subfamilyStr = parts[2]
		}
		vmSize.Subfamily = &subfamilyStr
	}

	// [# of vCPUs] - ([0-9]+): Captures one or more digits.
	vmSize.Cpus = parts[4]

	// [Constrained vCPUs]*
	// -?: Optionally captures a hyphen.
	// ((?:[0-9]+)?): Optionally captures another sequence of one or more digits.
	if len(parts[5]) > 0 {
		_, err := strconv.Atoi(parts[5])
		if err != nil {
			return nil, fmt.Errorf("converting constrained CPUs, %w", err)
		}
		vmSize.CpusConstrained = &parts[5]
	}

	// [Additive Features]
	// ((?:[abcdilmtspPr]+|C+|NP)?): Captures a sequence of letters representing certain attributes.
	// It can capture combinations like 'abcdilmtspPr' or 'C+' or 'NP'.
	vmSize.AdditiveFeatures = []rune(parts[6])

	// [Accelerator Type]*
	// _?: Optionally captures an underscore.
	// (?:([A-Z][0-9]+)_?)?: Optionally captures a pattern that starts with an uppercase letter followed by digits,
	// followed by an optional underscore.
	if len(parts[7]) > 0 {
		vmSize.AcceleratorType = &parts[7]
	}

	// [Confidential Child Capability]* - only AKS
	// (_cc_)?: Optionally captures the string "cc" with underscores on both sides.
	if parts[8] == "_cc_" {
		vmSize.ConfidentialChildCapability = true
	}

	// parts slice at index 8 disambiguates more enhanced memory and I/O capabilities
	// for Standard M memory-optimized VM series.
	// For example:
	// 1 in Standard_M96s_1_v3
	// and 2 in Standard_M96s_2_v3
	// Ref: https://learn.microsoft.com/en-us/azure/virtual-machines/msv3-mdsv3-medium-series

	// [MI300X]*
	// (_MI300X_)?: Optionally captures the string "_MI300X".
	// This is used to identify the MI300 series of VMs.
	if parts[10] == "MI300X" {
		vmSize.MI300Series = true
	}

	// [H100]*
	// (_H100_)?: Optionally captures the string "_H100".
	// This is used to identify the H100 series of VMs.
	if parts[11] == "H100" {
		vmSize.H100Series = true
	}

	// [Version]*
	// Optionally captures the pattern 'v' or 'V' followed by a digit from 1 to 9.
	vmSize.Version = parts[12]

	// [Promo]*
	// (_Promo)?: Optionally captures the string "_Promo".
	if parts[13] == "_Promo" {
		vmSize.PromoVersion = true
	}

	// [Series]
	subfamily := ""
	if vmSize.Subfamily != nil {
		subfamily = *vmSize.Subfamily
	}
	version := ""
	if len(vmSize.Version) > 0 {
		version = "_" + vmSize.Version
	}
	vmSize.Series = vmSize.Family + subfamily + string(vmSize.AdditiveFeatures) + version

	return &vmSize, nil
}
