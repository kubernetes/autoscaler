/*
Copyright 2015 The Kubernetes Authors.

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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"k8s.io/kubernetes/test/utils/image"

	// Never, ever remove the line with "/ginkgo". Without it,
	// the ginkgo test runner will not detect that this
	// directory contains a Ginkgo test suite.
	// See https://github.com/kubernetes/kubernetes/issues/74827
	// "github.com/onsi/ginkgo/v2"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/config"

	// define and freeze constants
	_ "k8s.io/kubernetes/test/e2e/feature"

	// reconfigure framework
	_ "k8s.io/kubernetes/test/e2e/framework/debug/init"
	_ "k8s.io/kubernetes/test/e2e/framework/metrics/init"
	_ "k8s.io/kubernetes/test/e2e/framework/node/init"
	_ "k8s.io/kubernetes/test/utils/format"
)

// handleFlags sets up all flags and parses the command line.
func handleFlags() {
	config.CopyFlags(config.Flags, flag.CommandLine)
	framework.RegisterCommonFlags(flag.CommandLine)
	framework.RegisterClusterFlags(flag.CommandLine)
	flag.Parse()
}

func TestMain(m *testing.M) {
	// Register test flags, then parse flags.
	handleFlags()

	if framework.TestContext.ListImages {
		for _, v := range image.GetImageConfigs() {
			fmt.Println(v.GetE2EImage())
		}
		os.Exit(0)
	}

	framework.AfterReadingAllFlags(&framework.TestContext)

	if flag.CommandLine.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unknown additional command line arguments: %s", flag.CommandLine.Args())
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}

var _ = ginkgo.ReportAfterEach(func(report ginkgo.SpecReport) {
	progressReporter.ProcessSpecReport(report)
})

var _ = ginkgo.ReportBeforeSuite(func(report ginkgo.Report) {
	progressReporter.SetTestsTotal(report.PreRunStats.SpecsThatWillRun)
})

var _ = ginkgo.ReportAfterSuite("Kubernetes e2e suite report", func(report ginkgo.Report) {
	var err error
	// The DetailsRepoerter will output details about every test (name, files, lines, etc) which helps
	// when documenting our tests.
	if len(framework.TestContext.SpecSummaryOutput) <= 0 {
		return
	}
	absPath, err := filepath.Abs(framework.TestContext.SpecSummaryOutput)
	if err != nil {
		klog.Errorf("%#v\n", err)
		panic(err)
	}
	f, err := os.Create(absPath)
	if err != nil {
		klog.Errorf("%#v\n", err)
		panic(err)
	}

	defer f.Close()

	for _, specReport := range report.SpecReports {
		b, err := specReport.MarshalJSON()
		if err != nil {
			klog.Errorf("Error in detail reporter: %v", err)
			return
		}
		_, err = f.Write(b)
		if err != nil {
			klog.Errorf("Error saving test details in detail reporter: %v", err)
			return
		}
		// Printing newline between records for easier viewing in various tools.
		_, err = fmt.Fprintln(f, "")
		if err != nil {
			klog.Errorf("Error saving test details in detail reporter: %v", err)
			return
		}
	}
})
