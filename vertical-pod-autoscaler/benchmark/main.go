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
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"

	"k8s.io/autoscaler/vertical-pod-autoscaler/benchmark/pkg/cluster"
	"k8s.io/autoscaler/vertical-pod-autoscaler/benchmark/pkg/component"
	"k8s.io/autoscaler/vertical-pod-autoscaler/benchmark/pkg/results"
)

func main() {
	var kubeconfig string
	var profilesFlag string
	var runs int
	var outputFile string
	var noisePercentage int

	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig (default: $KUBECONFIG or ~/.kube/config)")
	flag.StringVar(&profilesFlag, "profile", "small", "benchmark profiles (comma-separated): small,medium,large,xlarge,xxlarge")
	flag.IntVar(&runs, "runs", 1, "number of benchmark runs per profile for averaging")
	flag.StringVar(&outputFile, "output", "", "path to output CSV file (optional)")
	flag.IntVar(&noisePercentage, "noise-percentage", 0, "percentage of additional noise (unmanaged) ReplicaSets relative to managed ReplicaSets (0 = no noise) (optional)")
	klog.InitFlags(nil)
	flag.Parse()

	profileList := strings.Split(profilesFlag, ",")
	for _, p := range profileList {
		p = strings.TrimSpace(p)
		if _, ok := cluster.Profiles[p]; !ok {
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
	config.QPS = 200
	config.Burst = 400

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating kubernetes client: %v", err)
	}

	vpaClient, err := vpa_clientset.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating VPA client: %v", err)
	}

	ctx := context.Background()
	components := component.NewComponents(kubeClient, config)

	klog.Infof("=== VPA Benchmark with KWOK ===")
	klog.Infof("Profiles: %v, Runs per profile: %d", profileList, runs)

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: cluster.BenchmarkNamespace}}
	kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	profileResults := make(map[string]results.ComponentResults)
	allRunResults := make(map[string][]results.ComponentResults)

	for _, profile := range profileList {
		profile = strings.TrimSpace(profile)
		count := cluster.Profiles[profile]
		noiseCount := count * noisePercentage / 100

		if noiseCount > 0 {
			klog.Infof("========== Profile: %s (%d VPAs, %d noise RS, %d total pods) ==========",
				profile, count, noiseCount, (count+noiseCount)*cluster.ReplicasPerReplicaSet)
		} else {
			klog.Infof("========== Profile: %s (%d VPAs) ==========", profile, count)
		}

		allResults := make([]results.ComponentResults, 0, runs)

		for run := 1; run <= runs; run++ {
			if runs > 1 {
				klog.Infof("--- Run %d/%d ---", run, runs)
			}

			iterResults, err := runBenchmarkIteration(ctx, kubeClient, vpaClient, components, count, noiseCount)
			if err != nil {
				klog.Warningf("Run %d failed: %v", run, err)
				continue
			}

			allResults = append(allResults, iterResults)
		}

		if len(allResults) == 0 {
			klog.Warningf("No successful runs for profile %s!", profile)
			continue
		}

		allRunResults[profile] = allResults
		profileResults[profile] = results.AverageResults(allResults)
	}

	klog.Infof("Final cleanup...")
	cluster.CleanupBenchmarkResources(ctx, kubeClient, vpaClient)

	results.PrintRunSummary(profileList, allRunResults)

	if len(profileResults) > 0 {
		results.PrintResultsTable(profileList, profileResults, cluster.Profiles, outputFile, noisePercentage)
	}

	klog.Infof("Benchmark completed successfully.")
}

// runBenchmarkIteration runs a single benchmark iteration:
//  1. Scale down all VPA components, clean up previous resources.
//  2. Create ReplicaSets, VPAs, and wait for KWOK pods.
//  3. Scale up recommender + admission controller, wait for recommendations.
//  4. Scrape recommender latency metrics.
//  5. Scale up updater, scrape its latency metrics once its loop completes.
//  6. Scrape admission controller metrics (accumulated during updater evictions).
func runBenchmarkIteration(ctx context.Context, kubeClient kubernetes.Interface, vpaClient vpa_clientset.Interface, components *component.Components, count int, noiseCount int) (results.ComponentResults, error) {
	klog.Infof("Scaling down VPA components...")
	for _, c := range components.All {
		if err := c.ScaleDown(ctx); err != nil {
			return nil, err
		}
	}

	klog.Infof("Deleting all VPA checkpoints...")
	cluster.DeleteAllVPACheckpoints(ctx, vpaClient)

	klog.Infof("Cleaning up existing benchmark resources...")
	cluster.CleanupBenchmarkResources(ctx, kubeClient, vpaClient)

	klog.Infof("Creating %d ReplicaSets (%d pods each, %d total)...", count, cluster.ReplicasPerReplicaSet, count*cluster.ReplicasPerReplicaSet)
	cluster.CreateInParallel(ctx, count, func(ctx context.Context, name string) {
		cluster.CreateReplicaSet(ctx, kubeClient, name)
	})

	if noiseCount > 0 {
		klog.Infof("Creating %d noise ReplicaSets (%d pods each, %d noise pods)...",
			noiseCount, cluster.ReplicasPerReplicaSet, noiseCount*cluster.ReplicasPerReplicaSet)
		cluster.CreateNoiseReplicaSets(ctx, kubeClient, noiseCount)
	}

	klog.Infof("Creating %d VPAs...", count)
	cluster.CreateInParallel(ctx, count, func(ctx context.Context, name string) {
		cluster.CreateVPA(ctx, vpaClient, name)
	})

	expectedPods := (count + noiseCount) * cluster.ReplicasPerReplicaSet
	klog.Infof("Waiting for %d KWOK pods to be running (%d managed + %d noise)...",
		expectedPods, count*cluster.ReplicasPerReplicaSet, noiseCount*cluster.ReplicasPerReplicaSet)
	err := wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(cluster.BenchmarkNamespace).List(ctx, metav1.ListOptions{
			FieldSelector: "status.phase=Running",
		})
		klog.Infof("> Pods: %d/%d", len(pods.Items), expectedPods)
		return len(pods.Items) >= expectedPods, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timeout waiting for pods: %v", err)
	}

	klog.Infof("Scaling up recommender and admission controller...")
	for _, c := range []*component.Component{components.Recommender, components.Admission} {
		if err := c.ScaleUp(ctx); err != nil {
			return nil, err
		}
	}

	klog.Infof("Waiting for VPA recommendations...")
	wait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		vpas, _ := vpaClient.AutoscalingV1().VerticalPodAutoscalers(cluster.BenchmarkNamespace).List(ctx, metav1.ListOptions{})
		withRec := 0
		for _, v := range vpas.Items {
			if v.Status.Recommendation != nil {
				withRec++
			}
		}
		klog.Infof("> VPAs with recommendations: %d/%d", withRec, count)
		return withRec == count, nil
	})

	iterResults := make(results.ComponentResults)

	klog.Infof("Scraping recommender metrics...")
	recResults, stopRecommender, err := components.Recommender.Scrape(ctx)
	if err != nil {
		return nil, err
	}
	defer stopRecommender()
	iterResults["recommender"] = recResults

	klog.Infof("Scaling up updater...")
	if err := components.Updater.ScaleUp(ctx); err != nil {
		return nil, err
	}

	klog.Infof("Scraping updater metrics...")
	updaterResults, stopUpdater, err := components.Updater.Scrape(ctx)
	if err != nil {
		return nil, err
	}
	defer stopUpdater()
	iterResults["updater"] = updaterResults

	klog.Infof("Scraping admission controller metrics...")
	admResults, stopAdmission, err := components.Admission.Scrape(ctx)
	if err != nil {
		return nil, err
	}
	defer stopAdmission()
	iterResults["admission"] = admResults

	return iterResults, nil
}
