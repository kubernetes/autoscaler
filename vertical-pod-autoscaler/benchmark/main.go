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

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/sync/errgroup"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
)

const (
	benchmarkNamespace = "benchmark"
	kwokNodeName       = "kwok-node"
	vpaNamespace       = "kube-system"
)

const replicasPerReplicaSet = 2

// profiles defines number of ReplicaSets (each with replicasPerReplicaSet pods)
var profiles = map[string]int{
	"small":   25,   // 25 VPAs, 25 ReplicaSets, 50 pods
	"medium":  100,  // 100 VPAs, 100 ReplicaSets, 200 pods
	"large":   250,  // 250 VPAs, 250 ReplicaSets, 500 pods
	"xlarge":  500,  // 500 VPAs, 500 ReplicaSets, 1000 pods
	"xxlarge": 1000, // 1000 VPAs, 1000 ReplicaSets, 2000 pods
}

var restConfig *rest.Config

func main() {
	var kubeconfig string
	var profilesFlag string
	var runs int
	var outputFile string
	var noisePercentage int

	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig (default: $KUBECONFIG or ~/.kube/config)")
	flag.StringVar(&profilesFlag, "profile", "small", "benchmark profiles (comma-separated): small,medium,large,xlarge,xxlarge,huge")
	flag.IntVar(&runs, "runs", 1, "number of benchmark runs per profile for averaging")
	flag.StringVar(&outputFile, "output", "", "path to output CSV file (optional)")
	flag.IntVar(&noisePercentage, "noise-percentage", 0, "percentage of additional noise (unmanaged) ReplicaSets relative to managed ReplicaSets (0 = no noise) (optional)")
	klog.InitFlags(nil)
	flag.Parse()

	profileList := strings.Split(profilesFlag, ",")
	for _, p := range profileList {
		p = strings.TrimSpace(p)
		if _, ok := profiles[p]; !ok {
			klog.Fatalf("Unknown profile: %s", p)
		}
	}

	if noisePercentage < 0 {
		klog.Fatalf("noise-percentage must not be negative")
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}
	// Increase rate limits to avoid client-side throttling during benchmark
	config.QPS = 200
	config.Burst = 400
	restConfig = config

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating kubernetes client: %v", err)
	}

	vpaClient, err := vpa_clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating VPA client: %v", err)
	}

	ctx := context.Background()

	fmt.Printf("=== VPA Benchmark with KWOK ===\n")
	fmt.Printf("Profiles: %v, Runs per profile: %d\n\n", profileList, runs)

	// Ensure namespace
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: benchmarkNamespace}}
	kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	// Collect results for all profiles
	profileResults := make(map[string]map[string]float64)
	// Store all individual run results for summary
	allRunResults := make(map[string][]map[string]float64)

	for _, profile := range profileList {
		profile = strings.TrimSpace(profile)
		count := profiles[profile]
		noiseCount := count * noisePercentage / 100

		if noiseCount > 0 {
			fmt.Printf("\n========== Profile: %s (%d VPAs, %d noise RS, %d total pods) ==========\n",
				profile, count, noiseCount, (count+noiseCount)*replicasPerReplicaSet)
		} else {
			fmt.Printf("\n========== Profile: %s (%d VPAs) ==========\n", profile, count)
		}

		// Collect results from all runs for this profile
		allResults := make([]map[string]float64, 0, runs)

		for run := 1; run <= runs; run++ {
			if runs > 1 {
				fmt.Printf("\n--- Run %d/%d ---\n", run, runs)
			}

			// Run one benchmark iteration
			latencies, err := runBenchmarkIteration(ctx, kubeClient, vpaClient, count, noiseCount)
			if err != nil {
				klog.Warningf("Run %d failed: %v", run, err)
				continue
			}

			allResults = append(allResults, latencies)

			// Print this run's results
			if runs > 1 {
				printLatencies(latencies, fmt.Sprintf("Run %d Results", run))
			}
		}

		if len(allResults) == 0 {
			fmt.Printf("No successful runs for profile %s!\n", profile)
			continue
		}

		// Store all run results for summary
		allRunResults[profile] = allResults

		// Calculate average for this profile
		var avgLatencies map[string]float64
		if runs == 1 {
			avgLatencies = allResults[0]
		} else {
			avgLatencies = averageLatencies(allResults)
		}
		profileResults[profile] = avgLatencies

		printLatencies(avgLatencies, fmt.Sprintf("%s Results", profile))
	}

	// Final cleanup
	fmt.Println("\nFinal cleanup...")
	cleanupBenchmarkResources(ctx, kubeClient, vpaClient)

	// Print run summary if multiple runs
	if runs > 1 {
		printRunSummary(profileList, allRunResults)
	}

	// Print results table and write to file
	if len(profileResults) > 0 {
		printResultsTable(profileList, profileResults, outputFile, noisePercentage)
	}

	fmt.Println("\nBenchmark completed successfully.")
}

func runBenchmarkIteration(ctx context.Context, kubeClient kubernetes.Interface, vpaClient vpa_clientset.Interface, count int, noiseCount int) (map[string]float64, error) {
	// Step 1: Scale down VPA components
	fmt.Println("Scaling down VPA components...")
	if err := scaleDownVPAComponents(ctx, kubeClient); err != nil {
		return nil, fmt.Errorf("failed to scale down VPA components: %v", err)
	}

	// Step 2: Delete all VPA checkpoints
	fmt.Println("Deleting all VPA checkpoints...")
	deleteAllVPACheckpoints(ctx, vpaClient)

	// Step 3: Cleanup any existing benchmark resources
	fmt.Println("Cleaning up existing benchmark resources...")
	cleanupBenchmarkResources(ctx, kubeClient, vpaClient)

	// Step 4: Create ReplicaSets (targeting KWOK node)
	fmt.Printf("Creating %d ReplicaSets (%d pods each, %d total)...\n", count, replicasPerReplicaSet, count*replicasPerReplicaSet)
	createInParallel(ctx, count, func(ctx context.Context, name string) {
		rs := makeReplicaSet(name)
		err := withRetry(func() error {
			_, err := kubeClient.AppsV1().ReplicaSets(benchmarkNamespace).Create(ctx, rs, metav1.CreateOptions{})
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		})
		if err != nil {
			klog.Warningf("Error creating ReplicaSet %s: %v", name, err)
		}
	})

	// Step 4b: Create noise ReplicaSets (not managed by any VPA)
	if noiseCount > 0 {
		fmt.Printf("Creating %d noise ReplicaSets (%d pods each, %d noise pods)...\n",
			noiseCount, replicasPerReplicaSet, noiseCount*replicasPerReplicaSet)
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(50)
		for i := range noiseCount {
			g.Go(func() error {
				name := fmt.Sprintf("noise-%d", i)
				rs := makeReplicaSet(name)
				err := withRetry(func() error {
					_, err := kubeClient.AppsV1().ReplicaSets(benchmarkNamespace).Create(gctx, rs, metav1.CreateOptions{})
					if errors.IsAlreadyExists(err) {
						return nil
					}
					return err
				})
				if err != nil {
					klog.Warningf("Error creating noise ReplicaSet %s: %v", name, err)
				}
				return nil
			})
		}
		g.Wait()
	}

	// Step 5: Create VPAs (while VPA components are still down)
	fmt.Printf("Creating %d VPAs...\n", count)
	createInParallel(ctx, count, func(ctx context.Context, name string) {
		vpa := makeVPA(name)
		err := withRetry(func() error {
			_, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(benchmarkNamespace).Create(ctx, vpa, metav1.CreateOptions{})
			if errors.IsAlreadyExists(err) {
				return nil
			}
			return err
		})
		if err != nil {
			klog.Warningf("Error creating VPA %s: %v", name, err)
		}
	})

	// Step 6: Wait for pods to be running
	expectedPods := (count + noiseCount) * replicasPerReplicaSet
	fmt.Printf("Waiting for %d KWOK pods to be running (%d managed + %d noise)...\n",
		expectedPods, count*replicasPerReplicaSet, noiseCount*replicasPerReplicaSet)
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(benchmarkNamespace).List(ctx, metav1.ListOptions{
			FieldSelector: "status.phase=Running",
		})
		fmt.Printf("  Pods: %d/%d\n", len(pods.Items), expectedPods)
		return len(pods.Items) >= expectedPods, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timeout waiting for pods: %v", err)
	}

	// Step 7: Scale up recommender (not updater yet)
	fmt.Println("Scaling up recommender...")
	if err := scaleUpRecommender(ctx, kubeClient); err != nil {
		return nil, fmt.Errorf("failed to scale up recommender: %v", err)
	}

	// Step 8: Wait for VPA recommendations
	fmt.Println("Waiting for VPA recommendations...")
	wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		vpas, _ := vpaClient.AutoscalingV1().VerticalPodAutoscalers(benchmarkNamespace).List(ctx, metav1.ListOptions{})
		withRec := 0
		for _, v := range vpas.Items {
			if v.Status.Recommendation != nil {
				withRec++
			}
		}
		fmt.Printf("  VPAs with recommendations: %d/%d\n", withRec, count)
		return withRec == count, nil
	})

	// Step 9: Scale up updater (now that recommendations exist)
	fmt.Println("Scaling up updater...")
	if err := scaleUpUpdater(ctx, kubeClient); err != nil {
		return nil, fmt.Errorf("failed to scale up updater: %v", err)
	}

	// Step 10: Wait for updater's first loop
	// The updater uses time.Tick which waits the full interval before the first tick
	// We set --updater-interval=2m, so wait 2 minutes for the first loop to start
	fmt.Println("Waiting 2 minutes for updater's first loop...")
	time.Sleep(2 * time.Minute)

	// Step 11: Poll for updater metrics (until 'total' step appears)
	// this step appears because of the "defer timer.ObserveTotal()" in updater.RunOnce()
	fmt.Println("Polling for updater metrics...")
	return waitForAndScrapeMetrics(ctx, kubeClient)
}

func scaleDeployment(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string, replicas int32) error {
	return withRetry(func() error {
		scale, err := kubeClient.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		scale.Spec.Replicas = replicas
		_, err = kubeClient.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
		return err
	})
}

func scaleVPADeployments(ctx context.Context, kubeClient kubernetes.Interface, replicas int32) error {
	deployments := []string{"vpa-updater", "vpa-recommender", "vpa-admission-controller"}
	for _, name := range deployments {
		if err := scaleDeployment(ctx, kubeClient, vpaNamespace, name, replicas); err != nil {
			return err
		}
	}
	return nil
}

func scaleDownVPAComponents(ctx context.Context, kubeClient kubernetes.Interface) error {
	// Scale all VPA deployments to 0
	if err := scaleVPADeployments(ctx, kubeClient, 0); err != nil {
		return err
	}

	// Wait for pods to be gone
	wait.PollUntilContextTimeout(ctx, 2*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(vpaNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app in (vpa-updater,vpa-recommender,vpa-admission-controller)",
		})
		if len(pods.Items) == 0 {
			return true, nil
		}
		fmt.Printf("  Waiting for %d VPA pods to terminate...\n", len(pods.Items))
		return false, nil
	})
	fmt.Println("  VPA components scaled down")
	return nil
}

func scaleUpRecommender(ctx context.Context, kubeClient kubernetes.Interface) error {
	// Scale recommender and admission-controller to 1
	if err := scaleDeployment(ctx, kubeClient, vpaNamespace, "vpa-recommender", 1); err != nil {
		return err
	}
	if err := scaleDeployment(ctx, kubeClient, vpaNamespace, "vpa-admission-controller", 1); err != nil {
		return err
	}
	fmt.Println("  Waiting for recommender to be ready...")
	if err := waitForVPAPodReady(ctx, kubeClient, "vpa-recommender"); err != nil {
		return err
	}
	fmt.Println("  Recommender ready")
	return nil
}

func scaleUpUpdater(ctx context.Context, kubeClient kubernetes.Interface) error {
	if err := scaleDeployment(ctx, kubeClient, vpaNamespace, "vpa-updater", 1); err != nil {
		return err
	}
	fmt.Println("  Waiting for updater to be ready...")
	if err := waitForVPAPodReady(ctx, kubeClient, "vpa-updater"); err != nil {
		return err
	}
	fmt.Println("  Updater ready")
	return nil
}

func waitForVPAPodReady(ctx context.Context, kubeClient kubernetes.Interface, appLabel string) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 120*time.Second, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(vpaNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", appLabel),
		})
		for _, p := range pods.Items {
			if p.Status.Phase == corev1.PodRunning {
				allReady := true
				for _, c := range p.Status.ContainerStatuses {
					if !c.Ready {
						allReady = false
						break
					}
				}
				if allReady {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

func deleteAllVPACheckpoints(ctx context.Context, vpaClient vpa_clientset.Interface) {
	nsList, _ := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints("").List(ctx, metav1.ListOptions{})
	for _, cp := range nsList.Items {
		vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(cp.Namespace).Delete(ctx, cp.Name, metav1.DeleteOptions{})
	}
	fmt.Printf("  Deleted %d VPA checkpoints\n", len(nsList.Items))
}

func makeReplicaSet(name string) *appsv1.ReplicaSet {
	replicas := int32(replicasPerReplicaSet)
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: benchmarkNamespace},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					// Directly assign to KWOK node - bypasses scheduler for faster pod creation
					NodeName: kwokNodeName,
					// Tolerate the kwok node taint
					Tolerations: []corev1.Toleration{{
						Key:      "kwok.x-k8s.io/node",
						Operator: corev1.TolerationOpEqual,
						Value:    "fake",
						Effect:   corev1.TaintEffectNoSchedule,
					}},
					Containers: []corev1.Container{{
						Name:  "app",
						Image: "registry.k8s.io/pause:3.10",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("10Mi"),
							},
						},
					}},
				},
			},
		},
	}
}

func makeVPA(name string) *vpa_types.VerticalPodAutoscaler {
	// KWOK pods don't support in-place scaling right now, so we use Recreate
	mode := vpa_types.UpdateModeRecreate
	return &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: benchmarkNamespace},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "ReplicaSet",
				Name:       name,
			},
			UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &mode},
		},
	}
}

func cleanupBenchmarkResources(ctx context.Context, kubeClient kubernetes.Interface, vpaClient vpa_clientset.Interface) {
	vpaClient.AutoscalingV1().VerticalPodAutoscalers(benchmarkNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})

	kubeClient.AppsV1().ReplicaSets(benchmarkNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})

	wait.PollUntilContextTimeout(ctx, time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(benchmarkNamespace).List(ctx, metav1.ListOptions{})
		return len(pods.Items) == 0, nil
	})
}

func waitForAndScrapeMetrics(ctx context.Context, kubeClient kubernetes.Interface) (map[string]float64, error) {
	pods, err := kubeClient.CoreV1().Pods(vpaNamespace).List(ctx, metav1.ListOptions{LabelSelector: "app=vpa-updater"})
	if err != nil || len(pods.Items) == 0 {
		return nil, fmt.Errorf("no vpa-updater pod found")
	}

	// The port-forward is configured to listen on this arbitrary port
	port := 18943

	podName := pods.Items[0].Name
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	go func() {
		url := kubeClient.CoreV1().RESTClient().Post().
			Resource("pods").Namespace(vpaNamespace).Name(podName).
			SubResource("portforward").URL()

		transport, upgrader, _ := spdy.RoundTripperFor(restConfig)
		dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
		pf, _ := portforward.New(dialer, []string{fmt.Sprintf("%d:8943", port)}, stopChan, readyChan, io.Discard, io.Discard)
		pf.ForwardPorts()
	}()

	select {
	case <-readyChan:
	case <-time.After(10 * time.Second):
		close(stopChan)
		return nil, fmt.Errorf("port-forward timeout")
	}
	defer close(stopChan)

	// Poll metrics endpoint until 'total' step appears (indicates loop completion)
	// 3 minute timeout since we've already waited interval time for the loop to start
	var latencies map[string]float64
	startTime := time.Now()
	err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
		if err != nil {
			return false, nil // Keep trying
		}
		defer resp.Body.Close()

		latencies, err = parseMetrics(resp.Body)
		if err != nil {
			return false, nil
		}

		// Check if 'total' step is present - indicates loop completed
		if _, ok := latencies["total"]; ok {
			return true, nil
		}
		fmt.Printf("  Waiting for updater loop to complete. Elapsed: %.2fs\n", time.Since(startTime).Seconds())
		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for updater metrics: %v", err)
	}

	return latencies, nil
}

func parseMetrics(body io.Reader) (map[string]float64, error) {
	latencies := make(map[string]float64)
	re := regexp.MustCompile(`vpa_updater_execution_latency_seconds_sum\{step="([^"]+)"\}\s+([\d.e+-]+)`)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); len(m) == 3 {
			if v, err := strconv.ParseFloat(m[2], 64); err == nil {
				latencies[m[1]] = v
			}
		}
	}
	return latencies, nil
}

func printLatencies(latencies map[string]float64, title string) {
	fmt.Printf("\n=== %s ===\n", title)
	var steps []string
	for k := range latencies {
		steps = append(steps, k)
	}
	sort.Strings(steps)
	for _, s := range steps {
		fmt.Printf("  %-15s: %.4fs\n", s, latencies[s])
	}
}

func averageLatencies(results []map[string]float64) map[string]float64 {
	if len(results) == 0 {
		return nil
	}

	stepCounts := make(map[string]int)
	stepSums := make(map[string]float64)

	for _, r := range results {
		for step, val := range r {
			stepSums[step] += val
			stepCounts[step]++
		}
	}

	avg := make(map[string]float64)
	for step, sum := range stepSums {
		avg[step] = sum / float64(stepCounts[step])
	}
	return avg
}

func printRunSummary(profileList []string, allRunResults map[string][]map[string]float64) {
	for _, profile := range profileList {
		profile = strings.TrimSpace(profile)
		runResults, ok := allRunResults[profile]
		if !ok || len(runResults) == 0 {
			continue
		}

		// Get all steps
		stepSet := make(map[string]bool)
		for _, r := range runResults {
			for step := range r {
				stepSet[step] = true
			}
		}
		var steps []string
		for s := range stepSet {
			steps = append(steps, s)
		}
		sort.Strings(steps)

		// Build header: Step, Run1, Run2, ...
		header := []string{"Step"}
		for i := range runResults {
			header = append(header, fmt.Sprintf("Run %d", i+1))
		}

		// Build rows
		var rows [][]string
		for _, step := range steps {
			row := []string{step}
			for _, r := range runResults {
				if v, ok := r[step]; ok {
					row = append(row, fmt.Sprintf("%.4fs", v))
				} else {
					row = append(row, "-")
				}
			}
			rows = append(rows, row)
		}

		fmt.Printf("\n========== %s: All Runs ==========\n", profile)
		table := tablewriter.NewWriter(os.Stdout)
		table.Header(header)
		table.Bulk(rows)
		table.Render()
	}
}

func printResultsTable(profileList []string, results map[string]map[string]float64, outputFile string, noisePercentage int) {
	// Get all steps from metric results
	stepSet := make(map[string]bool)
	for _, r := range results {
		for step := range r {
			stepSet[step] = true
		}
	}
	var steps []string
	for s := range stepSet {
		steps = append(steps, s)
	}
	sort.Strings(steps)

	// Build header
	header := []string{"Step"}
	for _, p := range profileList {
		p = strings.TrimSpace(p)
		count := profiles[p]
		noiseCount := count * noisePercentage / 100
		if noiseCount > 0 {
			header = append(header, fmt.Sprintf("%s (%d+%dn)", p, count, noiseCount))
		} else {
			header = append(header, fmt.Sprintf("%s (%d)", p, count))
		}
	}

	// Build rows
	var rows [][]string
	for _, step := range steps {
		row := []string{step}
		for _, p := range profileList {
			p = strings.TrimSpace(p)
			if r, ok := results[p]; ok {
				if v, ok := r[step]; ok {
					row = append(row, fmt.Sprintf("%.4fs", v))
				} else {
					row = append(row, "-")
				}
			} else {
				row = append(row, "-")
			}
		}
		rows = append(rows, row)
	}

	fmt.Println("\n========== Results ==========")
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(header)
	table.Bulk(rows)
	table.Render()

	if outputFile != "" {
		var buf strings.Builder
		buf.WriteString(strings.Join(header, ",") + "\n")
		for _, row := range rows {
			buf.WriteString(strings.Join(row, ",") + "\n")
		}
		if err := os.WriteFile(outputFile, []byte(buf.String()), 0644); err != nil {
			klog.Warningf("Failed to write output file %s: %v", outputFile, err)
		} else {
			fmt.Printf("\nResults written to %s (CSV format)\n", outputFile)
		}
	}
}

func createInParallel(ctx context.Context, count int, createFn func(ctx context.Context, name string)) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(50)
	for i := range count {
		g.Go(func() error {
			name := fmt.Sprintf("bench-%d", i)
			createFn(gctx, name)
			return nil
		})
	}
	g.Wait()
}

// retryBackoff is used for transient API errors
var retryBackoff = wait.Backoff{
	Steps:    5,
	Duration: 100 * time.Millisecond,
	Factor:   2.0,
	Jitter:   0.1,
}

// withRetry wraps an API call with retry logic for transient errors
func withRetry(fn func() error) error {
	return retry.OnError(retryBackoff, func(err error) bool {
		// Retry on conflicts, server errors, and rate limiting
		return errors.IsConflict(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) || errors.IsServiceUnavailable(err)
	}, fn)
}
