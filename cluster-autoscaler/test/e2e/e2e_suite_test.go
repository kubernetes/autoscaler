//go:build e2e
// +build e2e

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

package e2e

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/third_party/helm"
	"sigs.k8s.io/e2e-framework/third_party/kind"
)

var (
	// projectImage is the name of the image which will be build and loaded
	// with the code source changes to be tested.
	projectImage = "cluster-autoscaler:dev"
	testEnv      env.Environment
	kwokRepoURL  = "https://kwok.sigs.k8s.io/charts/"
)

func TestMain(m *testing.M) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	klog.StartFlushDaemon(500 * time.Millisecond)

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	testEnv, err = env.NewWithContext(ctx, cfg)
	if err != nil {
		log.Fatalf("error creating test environment: %v", err)
	}

	kindClusterName := envconf.RandomName("e2e-cluster", 16)

	testEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.LoadDockerImageToCluster(kindClusterName, projectImage),

		InstallKwok(),
		InstallClusterAutoscaler(kindClusterName),

		envfuncs.SetupCRDs("../../apis/config/crd", "*"),
	)

	testEnv.Finish(
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			return context.WithoutCancel(ctx), nil
		},
		envfuncs.DestroyCluster(kindClusterName),
	)

	os.Exit(testEnv.Run(m))
}

// InstallKwok installs the KWOK controller and fast-stage policies statelessly via Helm
func InstallKwok() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		manager := helm.New(cfg.KubeconfigFile())

		// 1. Install the KWOK Controller
		err := manager.RunInstall(
			helm.WithName("kwok"),
			helm.WithNamespace("kube-system"),
			helm.WithChart("kwok"),
			helm.WithArgs(
				"--repo", kwokRepoURL,
				"--set", "hostNetwork=true",
			),
			helm.WithWait(),
		)
		if err != nil {
			return ctx, fmt.Errorf("failed to install kwok controller: %w", err)
		}

		// 2. Install the stage-fast CRs (Required so fake pods/nodes report as 'Ready')
		err = manager.RunInstall(
			helm.WithName("kwok-stage-fast"),
			helm.WithNamespace("kube-system"),
			helm.WithChart("stage-fast"),
			helm.WithArgs(
				"--repo", kwokRepoURL,
			),
			helm.WithWait(),
		)
		if err != nil {
			return ctx, fmt.Errorf("failed to install kwok-stage-fast policies: %w", err)
		}

		return ctx, nil
	}
}

func InstallClusterAutoscaler(clusterName string) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		manager := helm.New(config.KubeconfigFile())

		err := manager.RunInstall(
			helm.WithName("cluster-autoscaler"),
			helm.WithNamespace("kube-system"),
			helm.WithChart("../../charts/cluster-autoscaler"),
			helm.WithWait(),
			helm.WithArgs(
				"--set", "image.repository=cluster-autoscaler",
				"--set", "image.tag=dev",
				"--set", "image.pullPolicy=IfNotPresent",

				"--set", "tolerations[0].key=node-role.kubernetes.io/control-plane",
				"--set", "tolerations[0].operator=Exists",
				"--set", "tolerations[0].effect=NoSchedule",

				"--set", "cloudProvider=kwok",
				"--set", fmt.Sprintf("autoDiscovery.clusterName=kind-%s", clusterName),
			),
		)

		return ctx, err
	}
}
