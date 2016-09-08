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

package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	lastTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "cluster_autoscaler",
			Name:      "last_time_seconds",
			Help:      "Last time CA run some main loop fragment.",
		}, []string{"main"},
	)

	lastDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "cluster_autoscaler",
			Name:      "last_duration_microseconds",
			Help:      "Time spent in last main loop fragments in microseconds.",
		}, []string{"main"},
	)

	duration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "cluster_autoscaler",
			Name:      "duration_microseconds",
			Help:      "Time spent in main loop fragments in microseconds.",
		}, []string{"main"},
	)
)

func init() {
	prometheus.MustRegister(duration)
	prometheus.MustRegister(lastDuration)
	prometheus.MustRegister(lastTimestamp)
}

func durationToMicro(start time.Time) float64 {
	return float64(time.Now().Sub(start).Nanoseconds() / 1000)
}
