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

package gce

// PriceInfo is the interface to fetch the pricing information needed for gce pricing.
type PriceInfo interface {
	// BaseCpuPricePerHour gets the base cpu price per hour.
	BaseCpuPricePerHour() float64
	// BaseMemoryPricePerHourPerGb gets the base memory price per hour per Gb.
	BaseMemoryPricePerHourPerGb() float64
	// BasePreemptibleDiscount gets the base preemptible discount applicable.
	BasePreemptibleDiscount() float64
	// BaseGpuPricePerHour gets the base gpu price per hour.
	BaseGpuPricePerHour() float64

	// PredefinedCpuPricePerHour gets the predefined cpu price per hour for machine family.
	PredefinedCpuPricePerHour() map[string]float64
	// PredefinedMemoryPricePerHourPerGb gets the predefined memory price per hour per Gb for machine family.
	PredefinedMemoryPricePerHourPerGb() map[string]float64
	// PredefinedPreemptibleDiscount gets the predefined preemptible discount for machine family.
	PredefinedPreemptibleDiscount() map[string]float64

	// CustomCpuPricePerHour gets the cpu price per hour for custom machine of a machine family.
	CustomCpuPricePerHour() map[string]float64
	// CustomMemoryPricePerHourPerGb gets the memory price per hour per Gb for custom machine of a machine family.
	CustomMemoryPricePerHourPerGb() map[string]float64
	CustomPreemptibleDiscount() map[string]float64

	InstancePrices() map[string]float64
	PreemptibleInstancePrices() map[string]float64

	GpuPrices() map[string]float64
	// PreemptibleGpuPrices gets the price of preemptible GPUs.
	PreemptibleGpuPrices() map[string]float64

	// BootDiskPricePerHour returns the map of the prices per Gb of boot disk per hour.
	BootDiskPricePerHour() map[string]float64
	// LocalSsdPricePerHour returns the price per Gb of local SSD per hour.
	LocalSsdPricePerHour() float64
	// LocalSsdPricePerHour returns the price per Gb of local SSD per hour for Spot VMs.
	SpotLocalSsdPricePerHour() float64
}

const hoursInMonth = float64(24 * 30)

const (
	// TODO: Move it to a config file.
	cpuPricePerHour          = 0.033174
	memoryPricePerHourPerGb  = 0.004446
	preemptibleDiscount      = 0.00698 / 0.033174
	gpuPricePerHour          = 0.700
	localSsdPriceMonthly     = 0.08
	spotLocalSsdPriceMonthly = 0.048
)

var (
	predefinedCpuPricePerHour = map[string]float64{
		"a2":  0.031611,
		"c2":  0.03398,
		"c2d": 0.029563,
		"c3":  0.03398,
		"e2":  0.021811,
		"m1":  0.0348,
		"n1":  0.031611,
		"n2":  0.031611,
		"n2d": 0.027502,
		"t2d": 0.027502,
	}
	predefinedMemoryPricePerHourPerGb = map[string]float64{
		"a2":  0.004237,
		"c2":  0.00455,
		"c2d": 0.003959,
		"c3":  0.00456,
		"e2":  0.002923,
		"m1":  0.0051,
		"n1":  0.004237,
		"n2":  0.004237,
		"n2d": 0.003686,
		"t2d": 0.003686,
	}
	predefinedPreemptibleDiscount = map[string]float64{
		"a2":  0.009483 / 0.031611,
		"c2":  0.00822 / 0.03398,
		"c2d": 0.007154 / 0.029563,
		"c3":  0.003086 / 0.03398,
		"e2":  0.006543 / 0.021811,
		"m1":  0.00733 / 0.0348,
		"n1":  0.006655 / 0.031611,
		"n2":  0.007650 / 0.031611,
		"n2d": 0.002773 / 0.027502,
		"t2d": 0.006655 / 0.027502,
	}
	customCpuPricePerHour = map[string]float64{
		"e2":  0.022890,
		"n1":  0.033174,
		"n2":  0.033174,
		"n2d": 0.028877,
	}
	customMemoryPricePerHourPerGb = map[string]float64{
		"e2":  0.003067,
		"n1":  0.004446,
		"n2":  0.004446,
		"n2d": 0.003870,
	}
	customPreemptibleDiscount = map[string]float64{
		"e2":  0.006867 / 0.022890,
		"n1":  0.00698 / 0.033174,
		"n2":  0.00802 / 0.033174,
		"n2d": 0.002908 / 0.028877,
	}

	// e2-micro and e2-small have allocatable set too high resulting in
	// overcommit. To make cluster autoscaler prefer e2-medium given the choice
	// between the three machine types, the prices for e2-micro and e2-small
	// are artificially set to be higher than the price of e2-medium.
	instancePrices = map[string]float64{
		"a2-highgpu-1g":    3.673385,
		"a2-highgpu-2g":    7.34677,
		"a2-highgpu-4g":    14.69354,
		"a2-highgpu-8g":    29.38708,
		"a2-megagpu-16g":   55.739504,
		"a2-ultragpu-1g":   5.0688,
		"a2-ultragpu-2g":   10.1376,
		"a2-ultragpu-4g":   20.2752,
		"a2-ultragpu-8g":   40.5504,
		"c2-standard-4":    0.2088,
		"c2-standard-8":    0.4176,
		"c2-standard-16":   0.8352,
		"c2-standard-30":   1.5660,
		"c2-standard-60":   3.1321,
		"c2d-highcpu-2":    0.0750,
		"c2d-highcpu-4":    0.1499,
		"c2d-highcpu-8":    0.2998,
		"c2d-highcpu-16":   0.5997,
		"c2d-highcpu-32":   1.1994,
		"c2d-highcpu-56":   2.0989,
		"c2d-highcpu-112":  4.1979,
		"c2d-highmem-2":    0.1225,
		"c2d-highmem-4":    0.2449,
		"c2d-highmem-8":    0.4899,
		"c2d-highmem-16":   0.9798,
		"c2d-highmem-32":   1.9595,
		"c2d-highmem-56":   3.4292,
		"c2d-highmem-112":  6.8583,
		"c2d-standard-2":   0.0908,
		"c2d-standard-4":   0.1816,
		"c2d-standard-8":   0.3632,
		"c2d-standard-16":  0.7264,
		"c2d-standard-32":  1.4528,
		"c2d-standard-56":  2.5423,
		"c2d-standard-112": 5.0847,
		"c3-standard-4":    0.20888,
		"c3-standard-8":    0.41776,
		"c3-standard-22":   1.14884,
		"c3-standard-44":   2.29768,
		"c3-standard-88":   4.59536,
		"c3-standard-176":  9.19072,
		"c3-highmem-4":     0.28184,
		"c3-highmem-8":     0.56368,
		"c3-highmem-22":    1.55012,
		"c3-highmem-44":    3.10024,
		"c3-highmem-88":    6.20048,
		"c3-highmem-176":   12.40096,
		"c3-highcpu-4":     0.1724,
		"c3-highcpu-8":     0.3448,
		"c3-highcpu-22":    0.9482,
		"c3-highcpu-44":    1.8964,
		"c3-highcpu-88":    3.7928,
		"c3-highcpu-176":   7.5856,
		"e2-highcpu-2":     0.04947,
		"e2-highcpu-4":     0.09894,
		"e2-highcpu-8":     0.19788,
		"e2-highcpu-16":    0.39576,
		"e2-highcpu-32":    0.79149,
		"e2-highmem-2":     0.09040,
		"e2-highmem-4":     0.18080,
		"e2-highmem-8":     0.36160,
		"e2-highmem-16":    0.72320,
		"e2-medium":        0.03351,
		"e2-micro":         0.03353, // Should be 0.00838. Set to be > e2-medium.
		"e2-small":         0.03352, // Should be 0.01675. Set to be > e2-medium.
		"e2-standard-2":    0.06701,
		"e2-standard-4":    0.13402,
		"e2-standard-8":    0.26805,
		"e2-standard-16":   0.53609,
		"e2-standard-32":   1.07210,
		"f1-micro":         0.0076,
		"g1-small":         0.0257,
		"m1-megamem-96":    10.6740,
		"m1-ultramem-40":   6.3039,
		"m1-ultramem-80":   12.6078,
		"m1-ultramem-160":  25.2156,
		"m2-ultramem-208":  42.186,
		"m2-ultramem-416":  84.371,
		"m2-megamem-416":   50.372,
		"n1-highcpu-2":     0.0709,
		"n1-highcpu-4":     0.1418,
		"n1-highcpu-8":     0.2836,
		"n1-highcpu-16":    0.5672,
		"n1-highcpu-32":    1.1344,
		"n1-highcpu-64":    2.2688,
		"n1-highcpu-96":    3.402,
		"n1-highmem-2":     0.1184,
		"n1-highmem-4":     0.2368,
		"n1-highmem-8":     0.4736,
		"n1-highmem-16":    0.9472,
		"n1-highmem-32":    1.8944,
		"n1-highmem-64":    3.7888,
		"n1-highmem-96":    5.6832,
		"n1-standard-1":    0.0475,
		"n1-standard-2":    0.0950,
		"n1-standard-4":    0.1900,
		"n1-standard-8":    0.3800,
		"n1-standard-16":   0.7600,
		"n1-standard-32":   1.5200,
		"n1-standard-64":   3.0400,
		"n1-standard-96":   4.5600,
		"n2-highcpu-2":     0.0717,
		"n2-highcpu-4":     0.1434,
		"n2-highcpu-8":     0.2868,
		"n2-highcpu-16":    0.5736,
		"n2-highcpu-32":    1.1471,
		"n2-highcpu-48":    1.7207,
		"n2-highcpu-64":    2.2943,
		"n2-highcpu-80":    2.8678,
		"n2-highcpu-96":    3.4414,
		"n2-highcpu-128":   4.5886,
		"n2-highmem-2":     0.1310,
		"n2-highmem-4":     0.2620,
		"n2-highmem-8":     0.5241,
		"n2-highmem-16":    1.0481,
		"n2-highmem-32":    2.0962,
		"n2-highmem-48":    3.1443,
		"n2-highmem-64":    4.1924,
		"n2-highmem-80":    5.2406,
		"n2-highmem-96":    6.2886,
		"n2-highmem-128":   7.7069,
		"n2-standard-2":    0.0971,
		"n2-standard-4":    0.1942,
		"n2-standard-8":    0.3885,
		"n2-standard-16":   0.7769,
		"n2-standard-32":   1.5539,
		"n2-standard-48":   2.3308,
		"n2-standard-64":   3.1078,
		"n2-standard-80":   3.8847,
		"n2-standard-96":   4.6616,
		"n2-standard-128":  6.2156,
		"n2d-highcpu-2":    0.0624,
		"n2d-highcpu-4":    0.1248,
		"n2d-highcpu-8":    0.2495,
		"n2d-highcpu-16":   0.4990,
		"n2d-highcpu-32":   0.9980,
		"n2d-highcpu-48":   1.4970,
		"n2d-highcpu-64":   1.9960,
		"n2d-highcpu-80":   2.4950,
		"n2d-highcpu-96":   2.9940,
		"n2d-highcpu-128":  3.9920,
		"n2d-highcpu-224":  6.9861,
		"n2d-highmem-2":    0.1140,
		"n2d-highmem-4":    0.2280,
		"n2d-highmem-8":    0.4559,
		"n2d-highmem-16":   0.9119,
		"n2d-highmem-32":   1.8237,
		"n2d-highmem-48":   2.7356,
		"n2d-highmem-64":   3.6474,
		"n2d-highmem-80":   4.5593,
		"n2d-highmem-96":   5.4711,
		"n2d-standard-2":   0.0845,
		"n2d-standard-4":   0.1690,
		"n2d-standard-8":   0.3380,
		"n2d-standard-16":  0.6759,
		"n2d-standard-32":  1.3519,
		"n2d-standard-48":  2.0278,
		"n2d-standard-64":  2.7038,
		"n2d-standard-80":  3.3797,
		"n2d-standard-96":  4.0556,
		"n2d-standard-128": 5.4075,
		"n2d-standard-224": 9.4632,
		"t2d-standard-1":   0.0422,
		"t2d-standard-2":   0.0845,
		"t2d-standard-4":   0.1690,
		"t2d-standard-8":   0.3380,
		"t2d-standard-16":  0.6759,
		"t2d-standard-32":  1.3519,
		"t2d-standard-48":  2.0278,
		"t2d-standard-60":  2.5348,
	}
	preemptiblePrices = map[string]float64{
		"a2-highgpu-1g":    1.102016,
		"a2-highgpu-2g":    2.204031,
		"a2-highgpu-4g":    4.408062,
		"a2-highgpu-8g":    8.816124,
		"a2-megagpu-16g":   16.721851,
		"a2-ultragpu-1g":   1.6,
		"a2-ultragpu-2g":   3.2,
		"a2-ultragpu-4g":   6.4,
		"a2-ultragpu-8g":   12.8,
		"c2-standard-4":    0.0505,
		"c2-standard-8":    0.1011,
		"c2-standard-16":   0.2021,
		"c2-standard-30":   0.3790,
		"c2-standard-60":   0.7579,
		"c2d-highcpu-2":    0.0181,
		"c2d-highcpu-4":    0.0363,
		"c2d-highcpu-8":    0.0726,
		"c2d-highcpu-16":   0.1451,
		"c2d-highcpu-32":   0.2902,
		"c2d-highcpu-56":   0.5079,
		"c2d-highcpu-112":  1.0158,
		"c2d-highmem-2":    0.0296,
		"c2d-highmem-4":    0.0593,
		"c2d-highmem-8":    0.1185,
		"c2d-highmem-16":   0.2371,
		"c2d-highmem-32":   0.4742,
		"c2d-highmem-56":   0.8298,
		"c2d-highmem-112":  1.6596,
		"c2d-standard-2":   0.0220,
		"c2d-standard-4":   0.0439,
		"c2d-standard-8":   0.0879,
		"c2d-standard-16":  0.1758,
		"c2d-standard-32":  0.3516,
		"c2d-standard-56":  0.6152,
		"c2d-standard-112": 1.2304,
		"c3-standard-4":    0.018952,
		"c3-standard-8":    0.037904,
		"c3-standard-22":   0.104236,
		"c3-standard-44":   0.208472,
		"c3-standard-88":   0.416944,
		"c3-standard-176":  0.833888,
		"c3-highmem-4":     0.02556,
		"c3-highmem-8":     0.05112,
		"c3-highmem-22":    0.14058,
		"c3-highmem-44":    0.28116,
		"c3-highmem-88":    0.56232,
		"c3-highmem-176":   1.12464,
		"c3-highcpu-4":     0.015648,
		"c3-highcpu-8":     0.031296,
		"c3-highcpu-22":    0.086064,
		"c3-highcpu-44":    0.172128,
		"c3-highcpu-88":    0.344256,
		"c3-highcpu-176":   0.688512,
		"e2-highcpu-2":     0.01484,
		"e2-highcpu-4":     0.02968,
		"e2-highcpu-8":     0.05936,
		"e2-highcpu-16":    0.11873,
		"e2-highcpu-32":    0.23744,
		"e2-highmem-2":     0.02712,
		"e2-highmem-4":     0.05424,
		"e2-highmem-8":     0.10848,
		"e2-highmem-16":    0.21696,
		"e2-medium":        0.01005,
		"e2-micro":         0.01007, // Should be 0.00251. Set to be > e2-medium.
		"e2-small":         0.01006, // Should be 0.00503. Set to be > e2-medium.
		"e2-standard-2":    0.02010,
		"e2-standard-4":    0.04021,
		"e2-standard-8":    0.08041,
		"e2-standard-16":   0.16083,
		"e2-standard-32":   0.32163,
		"f1-micro":         0.0035,
		"g1-small":         0.0070,
		"m1-megamem-96":    2.2600,
		"m1-ultramem-40":   1.3311,
		"m1-ultramem-80":   2.6622,
		"m1-ultramem-160":  5.3244,
		"n1-highcpu-2":     0.0150,
		"n1-highcpu-4":     0.0300,
		"n1-highcpu-8":     0.0600,
		"n1-highcpu-16":    0.1200,
		"n1-highcpu-32":    0.2400,
		"n1-highcpu-64":    0.4800,
		"n1-highcpu-96":    0.7200,
		"n1-highmem-2":     0.0250,
		"n1-highmem-4":     0.0500,
		"n1-highmem-8":     0.1000,
		"n1-highmem-16":    0.2000,
		"n1-highmem-32":    0.4000,
		"n1-highmem-64":    0.8000,
		"n1-highmem-96":    1.2000,
		"n1-standard-1":    0.0100,
		"n1-standard-2":    0.0200,
		"n1-standard-4":    0.0400,
		"n1-standard-8":    0.0800,
		"n1-standard-16":   0.1600,
		"n1-standard-32":   0.3200,
		"n1-standard-64":   0.6400,
		"n1-standard-96":   0.9600,
		"n2-highcpu-2":     0.0173,
		"n2-highcpu-4":     0.0347,
		"n2-highcpu-8":     0.0694,
		"n2-highcpu-16":    0.1388,
		"n2-highcpu-32":    0.2776,
		"n2-highcpu-48":    0.4164,
		"n2-highcpu-64":    0.5552,
		"n2-highcpu-80":    0.6940,
		"n2-highcpu-96":    0.8328,
		"n2-highcpu-128":   1.1104,
		"n2-highmem-2":     0.0317,
		"n2-highmem-4":     0.0634,
		"n2-highmem-8":     0.1268,
		"n2-highmem-16":    0.2536,
		"n2-highmem-32":    0.5073,
		"n2-highmem-48":    0.7609,
		"n2-highmem-64":    1.0145,
		"n2-highmem-80":    1.2681,
		"n2-highmem-96":    1.5218,
		"n2-highmem-128":   1.8691,
		"n2-standard-2":    0.0235,
		"n2-standard-4":    0.0470,
		"n2-standard-8":    0.0940,
		"n2-standard-16":   0.1880,
		"n2-standard-32":   0.3760,
		"n2-standard-48":   0.5640,
		"n2-standard-64":   0.7520,
		"n2-standard-80":   0.9400,
		"n2-standard-96":   1.128,
		"n2-standard-128":  1.504,
		"n2d-highcpu-2":    0.00629,
		"n2d-highcpu-4":    0.01258,
		"n2d-highcpu-8":    0.02516,
		"n2d-highcpu-16":   0.05032,
		"n2d-highcpu-32":   0.10064,
		"n2d-highcpu-48":   0.15096,
		"n2d-highcpu-64":   0.20128,
		"n2d-highcpu-80":   0.2516,
		"n2d-highcpu-96":   0.30192,
		"n2d-highcpu-128":  0.40256,
		"n2d-highcpu-224":  0.70448,
		"n2d-highmem-2":    0.011498,
		"n2d-highmem-4":    0.022996,
		"n2d-highmem-8":    0.045992,
		"n2d-highmem-16":   0.091984,
		"n2d-highmem-32":   0.183968,
		"n2d-highmem-48":   0.275952,
		"n2d-highmem-64":   0.367936,
		"n2d-highmem-80":   0.45992,
		"n2d-highmem-96":   0.551904,
		"n2d-standard-2":   0.008522,
		"n2d-standard-4":   0.017044,
		"n2d-standard-8":   0.034088,
		"n2d-standard-16":  0.068176,
		"n2d-standard-32":  0.136352,
		"n2d-standard-48":  0.204528,
		"n2d-standard-64":  0.272704,
		"n2d-standard-80":  0.34088,
		"n2d-standard-96":  0.409056,
		"n2d-standard-128": 0.545408,
		"n2d-standard-224": 0.954464,
		"t2d-standard-1":   0.0102,
		"t2d-standard-2":   0.0204,
		"t2d-standard-4":   0.0409,
		"t2d-standard-8":   0.0818,
		"t2d-standard-16":  0.1636,
		"t2d-standard-32":  0.3271,
		"t2d-standard-48":  0.4907,
		"t2d-standard-60":  0.6134,
	}
	gpuPrices = map[string]float64{
		"nvidia-tesla-t4":   0.35,
		"nvidia-tesla-p4":   0.60,
		"nvidia-tesla-v100": 2.48,
		"nvidia-tesla-p100": 1.46,
		"nvidia-tesla-k80":  0.45,
		"nvidia-tesla-a100": 0, // price of this gpu is counted into A2 machine-type price
		"nvidia-a100-80gb":  0, // price of this gpu is counted into A2 machine-type price
	}
	preemptibleGpuPrices = map[string]float64{
		"nvidia-tesla-t4":   0.11,
		"nvidia-tesla-p4":   0.216,
		"nvidia-tesla-v100": 0.74,
		"nvidia-tesla-p100": 0.43,
		"nvidia-tesla-k80":  0.037500,
		"nvidia-tesla-a100": 0, // price of this gpu is counted into A2 machine-type price
		"nvidia-a100-80gb":  0, // price of this gpu is counted into A2 machine-type price
	}
	bootDiskPricePerHour = map[string]float64{
		"pd-standard": 0.04 / hoursInMonth,
		"pd-balanced": 0.100 / hoursInMonth,
		"pd-ssd":      0.170 / hoursInMonth,
	}
	// DefaultBootDiskType is pd-standard disk type.
	DefaultBootDiskType = "pd-standard"
)

// GcePriceInfo is the GCE specific implementation of the PricingInfo.
type GcePriceInfo struct {
	baseCpuPricePerHour         float64
	baseMemoryPricePerHourPerGb float64
	basePreemptibleDiscount     float64
	baseGpuPricePerHour         float64
	localSsdPriceMonthly        float64
	spotLocalSsdPriceMonthly    float64

	predefinedCpuPricePerHour         map[string]float64
	predefinedMemoryPricePerHourPerGb map[string]float64
	predefinedPreemptibleDiscount     map[string]float64

	customCpuPricePerHour         map[string]float64
	customMemoryPricePerHourPerGb map[string]float64
	customPreemptibleDiscount     map[string]float64

	instancePrices            map[string]float64
	preemptibleInstancePrices map[string]float64

	gpuPrices            map[string]float64
	preemptibleGpuPrices map[string]float64
	bootDiskPricePerHour map[string]float64
}

// NewGcePriceInfo returns a new instance of the GcePriceInfo.
func NewGcePriceInfo() *GcePriceInfo {
	return &GcePriceInfo{
		baseCpuPricePerHour:         cpuPricePerHour,
		baseMemoryPricePerHourPerGb: memoryPricePerHourPerGb,
		basePreemptibleDiscount:     preemptibleDiscount,
		baseGpuPricePerHour:         gpuPricePerHour,
		localSsdPriceMonthly:        localSsdPriceMonthly,
		spotLocalSsdPriceMonthly:    spotLocalSsdPriceMonthly,

		predefinedCpuPricePerHour:         predefinedCpuPricePerHour,
		predefinedMemoryPricePerHourPerGb: predefinedMemoryPricePerHourPerGb,
		predefinedPreemptibleDiscount:     predefinedPreemptibleDiscount,

		customCpuPricePerHour:         customCpuPricePerHour,
		customMemoryPricePerHourPerGb: customMemoryPricePerHourPerGb,
		customPreemptibleDiscount:     customPreemptibleDiscount,

		instancePrices:            instancePrices,
		preemptibleInstancePrices: preemptiblePrices,

		gpuPrices:            gpuPrices,
		preemptibleGpuPrices: preemptibleGpuPrices,
		bootDiskPricePerHour: bootDiskPricePerHour,
	}
}

// BaseCpuPricePerHour gets the base cpu price per hour
func (g *GcePriceInfo) BaseCpuPricePerHour() float64 {
	return g.baseCpuPricePerHour
}

// BaseMemoryPricePerHourPerGb gets the base memory price per hour per Gb
func (g *GcePriceInfo) BaseMemoryPricePerHourPerGb() float64 {
	return g.baseMemoryPricePerHourPerGb
}

// BasePreemptibleDiscount gets the base preemptible discount applicable
func (g *GcePriceInfo) BasePreemptibleDiscount() float64 {
	return g.basePreemptibleDiscount
}

// BaseGpuPricePerHour gets the base gpu price per hour
func (g *GcePriceInfo) BaseGpuPricePerHour() float64 {
	return g.baseGpuPricePerHour
}

// PredefinedCpuPricePerHour gets the predefined cpu price per hour for machine family
func (g *GcePriceInfo) PredefinedCpuPricePerHour() map[string]float64 {
	return g.predefinedCpuPricePerHour
}

// PredefinedMemoryPricePerHourPerGb gets the predefined memory price per hour per Gb for machine family
func (g *GcePriceInfo) PredefinedMemoryPricePerHourPerGb() map[string]float64 {
	return g.predefinedMemoryPricePerHourPerGb
}

// PredefinedPreemptibleDiscount gets the predefined preemptible discount for machine family
func (g *GcePriceInfo) PredefinedPreemptibleDiscount() map[string]float64 {
	return g.predefinedPreemptibleDiscount
}

// CustomCpuPricePerHour gets the cpu price per hour for custom machine of a machine family
func (g *GcePriceInfo) CustomCpuPricePerHour() map[string]float64 {
	return g.customCpuPricePerHour
}

// CustomMemoryPricePerHourPerGb gets the memory price per hour per Gb for custom machine of a machine family
func (g *GcePriceInfo) CustomMemoryPricePerHourPerGb() map[string]float64 {
	return g.customMemoryPricePerHourPerGb
}

// CustomPreemptibleDiscount gets the preemptible discount of a machine family
func (g *GcePriceInfo) CustomPreemptibleDiscount() map[string]float64 {
	return g.customPreemptibleDiscount
}

// InstancePrices gets the prices for standard machine types
func (g *GcePriceInfo) InstancePrices() map[string]float64 {
	return g.instancePrices
}

// PreemptibleInstancePrices gets the preemptible prices for standard machine types
func (g *GcePriceInfo) PreemptibleInstancePrices() map[string]float64 {
	return g.preemptibleInstancePrices
}

// GpuPrices gets the price of GPUs
func (g *GcePriceInfo) GpuPrices() map[string]float64 {
	return g.gpuPrices
}

// PreemptibleGpuPrices gets the price of preemptible GPUs
func (g *GcePriceInfo) PreemptibleGpuPrices() map[string]float64 {
	return g.preemptibleGpuPrices
}

// BootDiskPricePerHour gets the price of boot disk.
func (g *GcePriceInfo) BootDiskPricePerHour() map[string]float64 {
	return g.bootDiskPricePerHour
}

// LocalSsdPricePerHour gets the price of boot disk.
func (g *GcePriceInfo) LocalSsdPricePerHour() float64 {
	return g.localSsdPriceMonthly / hoursInMonth
}

// SpotLocalSsdPricePerHour gets the price of boot disk.
func (g *GcePriceInfo) SpotLocalSsdPricePerHour() float64 {
	return g.spotLocalSsdPriceMonthly / hoursInMonth
}
