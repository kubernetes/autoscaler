/*
Copyright The Kubernetes Authors.

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

package component

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/benchmark/pkg/cluster"
)

const (
	scaleDownPollInterval = 2 * time.Second
	scaleDownTimeout      = 60 * time.Second
	portForwardTimeout    = 10 * time.Second
	scrapePollInterval    = 10 * time.Second
	scrapeTimeout         = 10 * time.Minute
)

// Component represents a VPA component's full lifecycle: scaling, port-forwarding,
// and metrics scraping. Instantiated once via NewComponents with shared clients.
type Component struct {
	Label        string
	MetricPrefix string
	PodPort      int
	LocalPort    int
	PerRequest   bool

	kubeClient kubernetes.Interface
	restConfig *rest.Config
}

// Components holds the three VPA component instances and provides
// named access as well as iteration via All.
type Components struct {
	Recommender *Component
	Updater     *Component
	Admission   *Component
	All         []*Component
}

// NewComponents creates all VPA component instances with shared clients.
func NewComponents(kubeClient kubernetes.Interface, restConfig *rest.Config) *Components {
	c := &Components{
		Recommender: &Component{
			Label:        "recommender",
			MetricPrefix: "vpa_recommender",
			PodPort:      8942,
			LocalPort:    18942,
			kubeClient:   kubeClient,
			restConfig:   restConfig,
		},
		Updater: &Component{
			Label:        "updater",
			MetricPrefix: "vpa_updater",
			PodPort:      8943,
			LocalPort:    18943,
			kubeClient:   kubeClient,
			restConfig:   restConfig,
		},
		Admission: &Component{
			Label:        "admission-controller",
			MetricPrefix: "vpa_admission_controller",
			PodPort:      8944,
			LocalPort:    18944,
			PerRequest:   true,
			kubeClient:   kubeClient,
			restConfig:   restConfig,
		},
	}
	c.All = []*Component{c.Recommender, c.Updater, c.Admission}
	return c
}

// DeploymentName returns the Kubernetes deployment name for this component.
func (c *Component) DeploymentName() string {
	return "vpa-" + c.Label
}

// MetricsURL returns the local URL for this component's Prometheus metrics endpoint.
func (c *Component) MetricsURL() string {
	return fmt.Sprintf("http://localhost:%d/metrics", c.LocalPort)
}

// ScaleUp scales the component's deployment to 1 replica and waits for the pod
// to become ready.
func (c *Component) ScaleUp(ctx context.Context) error {
	if err := cluster.ScaleDeployment(ctx, c.kubeClient, cluster.VPANamespace, c.DeploymentName(), 1); err != nil {
		return fmt.Errorf("failed to scale up %s: %v", c.Label, err)
	}
	if err := cluster.WaitForVPAPodReady(ctx, c.kubeClient, c.Label); err != nil {
		return fmt.Errorf("%s not ready: %v", c.Label, err)
	}
	return nil
}

// ScaleDown scales the component's deployment to 0 replicas and waits for
// all its pods to terminate.
func (c *Component) ScaleDown(ctx context.Context) error {
	if err := cluster.ScaleDeployment(ctx, c.kubeClient, cluster.VPANamespace, c.DeploymentName(), 0); err != nil {
		return fmt.Errorf("failed to scale down %s: %v", c.Label, err)
	}
	return wait.PollUntilContextTimeout(ctx, scaleDownPollInterval, scaleDownTimeout, true, func(ctx context.Context) (bool, error) {
		pods, _ := c.kubeClient.CoreV1().Pods(cluster.VPANamespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/component=%s", c.Label),
		})
		return len(pods.Items) == 0, nil
	})
}

// PortForward sets up port forwarding to this component's pod.
// Returns a stop function to close the port-forward.
func (c *Component) PortForward(ctx context.Context) (stop func(), err error) {
	pods, err := c.kubeClient.CoreV1().Pods(cluster.VPANamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/component=%s", c.Label),
	})
	if err != nil || len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pod found for component %s", c.Label)
	}

	podName := pods.Items[0].Name
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	go func() {
		url := c.kubeClient.CoreV1().RESTClient().Post().
			Resource("pods").Namespace(cluster.VPANamespace).Name(podName).
			SubResource("portforward").URL()

		transport, upgrader, _ := spdy.RoundTripperFor(c.restConfig)
		dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
		pf, _ := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", c.LocalPort, c.PodPort)}, stopChan, readyChan, io.Discard, io.Discard)
		pf.ForwardPorts()
	}()

	select {
	case <-readyChan:
	case <-time.After(portForwardTimeout):
		close(stopChan)
		return nil, fmt.Errorf("port-forward timeout for %s", c.Label)
	}

	return func() { close(stopChan) }, nil
}

// ScrapeLatencies fetches /metrics from this component and parses execution latency.
func (c *Component) ScrapeLatencies() (sums, counts map[string]float64, err error) {
	resp, err := http.Get(c.MetricsURL())
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	return ParseExecutionLatency(resp.Body, c.MetricPrefix)
}

// ParseExecutionLatency parses Prometheus text format for execution_latency_seconds
// _sum and _count values, keyed by step label.
func ParseExecutionLatency(body io.Reader, metricPrefix string) (sums, counts map[string]float64, err error) {
	sums = make(map[string]float64)
	counts = make(map[string]float64)
	reSum := regexp.MustCompile(regexp.QuoteMeta(metricPrefix) + `_execution_latency_seconds_sum\{step="([^"]+)"\}\s+([\d.e+-]+)`)
	reCount := regexp.MustCompile(regexp.QuoteMeta(metricPrefix) + `_execution_latency_seconds_count\{step="([^"]+)"\}\s+([\d.e+-]+)`)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if m := reSum.FindStringSubmatch(line); len(m) == 3 {
			if v, err := strconv.ParseFloat(m[2], 64); err == nil {
				sums[m[1]] = v
			}
		}
		if m := reCount.FindStringSubmatch(line); len(m) == 3 {
			if v, err := strconv.ParseFloat(m[2], 64); err == nil {
				counts[m[1]] = v
			}
		}
	}
	return sums, counts, scanner.Err()
}

// Scrape port-forwards to this component, polls until the 'total' latency step
// appears, then returns the processed step map. For PerRequest components
// (admission controller) values are divided by invocation count. The returned
// stop function must be deferred by the caller to close the port-forward.
func (c *Component) Scrape(ctx context.Context) (stepResults map[string]float64, stop func(), err error) {
	stop, err = c.PortForward(ctx)
	if err != nil {
		return nil, nil, err
	}

	var sums, counts map[string]float64
	startTime := time.Now()
	pollErr := wait.PollUntilContextTimeout(ctx, scrapePollInterval, scrapeTimeout, true, func(ctx context.Context) (bool, error) {
		sums, counts, err = c.ScrapeLatencies()
		if err != nil {
			return false, nil
		}
		if _, ok := sums["total"]; ok {
			return true, nil
		}
		klog.Infof("> Waiting for %s metrics. Elapsed: %.2fs", c.Label, time.Since(startTime).Seconds())
		return false, nil
	})
	if pollErr != nil {
		stop()
		return nil, nil, fmt.Errorf("timed out waiting for %s metrics: %v", c.Label, pollErr)
	}

	if c.PerRequest {
		stepResults = make(map[string]float64)
		for step, sum := range sums {
			if cnt, ok := counts[step]; ok && cnt > 0 {
				stepResults[step] = sum / cnt
			}
		}
		if cnt, ok := counts["total"]; ok {
			stepResults["request_count"] = cnt
		}
	} else {
		stepResults = sums
	}

	return stepResults, stop, nil
}
