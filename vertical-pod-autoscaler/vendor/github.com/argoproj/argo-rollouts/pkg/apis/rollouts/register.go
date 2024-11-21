/*
Copyright 2017 The Kubernetes Authors.

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

package rollouts

const (
	Group string = "argoproj.io"

	RolloutKind     string = "Rollout"
	RolloutSingular string = "rollout"
	RolloutPlural   string = "rollouts"
	RolloutFullName string = RolloutPlural + "." + Group

	ExperimentKind     string = "Experiment"
	ExperimentSingular string = "experiment"
	ExperimentPlural   string = "experiments"
	ExperimentFullName string = ExperimentPlural + "." + Group

	AnalysisTemplateKind     string = "AnalysisTemplate"
	AnalysisTemplateSingular string = "analysistemplate"
	AnalysisTemplatePlural   string = "analysistemplates"
	AnalysisTemplateFullName string = AnalysisTemplatePlural + "." + Group

	ClusterAnalysisTemplateKind     string = "ClusterAnalysisTemplate"
	ClusterAnalysisTemplateSingular string = "clusteranalysistemplate"
	ClusterAnalysisTemplatePlural   string = "clusteranalysistemplates"
	ClusterAnalysisTemplateFullName string = ClusterAnalysisTemplatePlural + "." + Group

	AnalysisRunKind     string = "AnalysisRun"
	AnalysisRunSingular string = "analysisrun"
	AnalysisRunPlural   string = "analysisruns"
	AnalysisRunFullName string = AnalysisRunPlural + "." + Group
)
