//go:build e2e

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
)

var (
	testEnv env.Environment
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

	namespace := envconf.RandomName("testns", 12)

	testEnv.Setup(
		envfuncs.CreateNamespace(namespace),
	)

	testEnv.Finish(
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			return context.WithoutCancel(ctx), nil
		},
	)

	os.Exit(testEnv.Run(m))
}
