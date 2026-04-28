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

package results

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"k8s.io/klog/v2"
)

// ComponentResults maps component name -> step name -> value.
type ComponentResults = map[string]map[string]float64

// ComponentOrder defines the display order for component results.
var ComponentOrder = []string{"recommender", "updater", "admission"}

// ComponentTitles maps component keys to human-readable display titles.
var ComponentTitles = map[string]string{
	"recommender": "Recommender",
	"updater":     "Updater",
	"admission":   "Admission Controller",
}

// stepOrder defines execution order per component.
// Steps not listed here are appended alphabetically at the end.
var stepOrder = map[string][]string{
	"recommender": {"LoadVPAs", "LoadPods", "LoadMetrics", "UpdateVPAs", "MaintainCheckpoints", "GarbageCollect", "total"},
	"updater":     {"ListVPAs", "ListPods", "FilterPods", "AdmissionInit", "EvictPods", "total"},
	"admission":   {"read_request", "admit", "build_response", "write_response", "request_count", "total"},
}

func sortSteps(component string, steps []string) []string {
	order, ok := stepOrder[component]
	if !ok {
		sort.Strings(steps)
		return steps
	}
	rank := make(map[string]int, len(order))
	for i, s := range order {
		rank[s] = i
	}
	sort.Slice(steps, func(i, j int) bool {
		ri, oki := rank[steps[i]]
		rj, okj := rank[steps[j]]
		if oki && okj {
			return ri < rj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return steps[i] < steps[j]
	})
	return steps
}

func isCountField(step string) bool {
	return step == "request_count"
}

func formatValue(step string, v float64) string {
	if isCountField(step) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.4fs", v)
}

// PrintAllComponentLatencies prints per-component latency tables for a single result set.
func PrintAllComponentLatencies(results ComponentResults, title string) {
	for _, component := range ComponentOrder {
		latencies, ok := results[component]
		if !ok || len(latencies) == 0 {
			continue
		}
		componentTitle, ok := ComponentTitles[component]
		if !ok {
			componentTitle = component
		}
		printLatencies(component, latencies, fmt.Sprintf("%s [%s]", title, componentTitle))
	}
}

func printLatencies(component string, latencies map[string]float64, title string) {
	fmt.Printf("\n=== %s ===\n", title)
	var steps []string
	for k := range latencies {
		steps = append(steps, k)
	}
	sortSteps(component, steps)
	for _, s := range steps {
		fmt.Printf("  %-25s: %s\n", s, formatValue(s, latencies[s]))
	}
}

// AverageResults averages per-component, per-step values across multiple runs.
func AverageResults(results []ComponentResults) ComponentResults {
	if len(results) == 0 {
		return nil
	}

	type accumulator struct {
		sum   float64
		count int
	}
	accum := make(map[string]map[string]*accumulator)

	for _, r := range results {
		for component, steps := range r {
			if accum[component] == nil {
				accum[component] = make(map[string]*accumulator)
			}
			for step, val := range steps {
				if accum[component][step] == nil {
					accum[component][step] = &accumulator{}
				}
				accum[component][step].sum += val
				accum[component][step].count++
			}
		}
	}

	avg := make(map[string]map[string]float64)
	for component, steps := range accum {
		avg[component] = make(map[string]float64)
		for step, a := range steps {
			avg[component][step] = a.sum / float64(a.count)
		}
	}
	return avg
}

// PrintRunSummary prints a per-run comparison table for each component within each profile.
func PrintRunSummary(profileList []string, allRunResults map[string][]ComponentResults) {
	for _, profile := range profileList {
		profile = strings.TrimSpace(profile)
		runResults, ok := allRunResults[profile]
		if !ok || len(runResults) == 0 {
			continue
		}

		for _, component := range ComponentOrder {
			stepSet := make(map[string]bool)
			hasData := false
			for _, r := range runResults {
				if steps, ok := r[component]; ok {
					for step := range steps {
						stepSet[step] = true
					}
					hasData = true
				}
			}
			if !hasData {
				continue
			}

			var steps []string
			for s := range stepSet {
				steps = append(steps, s)
			}
			sortSteps(component, steps)

			header := []string{"Step"}
			for i := range runResults {
				header = append(header, fmt.Sprintf("Run %d", i+1))
			}
			header = append(header, "Avg")

			var rows [][]string
			for _, step := range steps {
				row := []string{step}
				var sum float64
				var count int
				for _, r := range runResults {
					if componentSteps, ok := r[component]; ok {
						if v, ok := componentSteps[step]; ok {
							row = append(row, formatValue(step, v))
							sum += v
							count++
						} else {
							row = append(row, "-")
						}
					} else {
						row = append(row, "-")
					}
				}
				if count > 0 {
					row = append(row, formatValue(step, sum/float64(count)))
				} else {
					row = append(row, "-")
				}
				rows = append(rows, row)
			}

			componentTitle := ComponentTitles[component]
			fmt.Printf("\n========== %s: All Runs [%s] ==========\n", profile, componentTitle)
			table := tablewriter.NewWriter(os.Stdout)
			table.Header(header)
			table.Bulk(rows)
			table.Render()
		}
	}
}

// PrintResultsTable prints a cross-profile comparison table for each component
// and optionally writes CSV output.
func PrintResultsTable(profileList []string, results map[string]ComponentResults, profiles map[string]int, outputFile string, noisePercentage int) {
	profileHeaders := make([]string, 0, len(profileList))
	for _, p := range profileList {
		p = strings.TrimSpace(p)
		count := profiles[p]
		noiseCount := count * noisePercentage / 100
		if noiseCount > 0 {
			profileHeaders = append(profileHeaders, fmt.Sprintf("%s (%d+%dn)", p, count, noiseCount))
		} else {
			profileHeaders = append(profileHeaders, fmt.Sprintf("%s (%d)", p, count))
		}
	}

	var csvBuf strings.Builder

	for _, component := range ComponentOrder {
		stepSet := make(map[string]bool)
		hasData := false
		for _, r := range results {
			if steps, ok := r[component]; ok {
				for step := range steps {
					stepSet[step] = true
				}
				hasData = true
			}
		}
		if !hasData {
			continue
		}

		var steps []string
		for s := range stepSet {
			steps = append(steps, s)
		}
		sortSteps(component, steps)

		header := append([]string{"Step"}, profileHeaders...)

		var rows [][]string
		for _, step := range steps {
			row := []string{step}
			for _, p := range profileList {
				p = strings.TrimSpace(p)
				if r, ok := results[p]; ok {
					if componentSteps, ok := r[component]; ok {
						if v, ok := componentSteps[step]; ok {
							row = append(row, formatValue(step, v))
						} else {
							row = append(row, "-")
						}
					} else {
						row = append(row, "-")
					}
				} else {
					row = append(row, "-")
				}
			}
			rows = append(rows, row)
		}

		componentTitle := ComponentTitles[component]
		fmt.Printf("\n========== Results [%s] ==========\n", componentTitle)
		table := tablewriter.NewWriter(os.Stdout)
		table.Header(header)
		table.Bulk(rows)
		table.Render()

		if outputFile != "" {
			csvBuf.WriteString(fmt.Sprintf("# %s\n", componentTitle))
			csvBuf.WriteString(strings.Join(header, ",") + "\n")
			for _, row := range rows {
				csvBuf.WriteString(strings.Join(row, ",") + "\n")
			}
			csvBuf.WriteString("\n")
		}
	}

	if outputFile != "" && csvBuf.Len() > 0 {
		if err := os.WriteFile(outputFile, []byte(csvBuf.String()), 0644); err != nil {
			klog.Warningf("Failed to write output file %s: %v", outputFile, err)
		} else {
			fmt.Printf("\nResults written to %s (CSV format)\n", outputFile)
		}
	}
}
