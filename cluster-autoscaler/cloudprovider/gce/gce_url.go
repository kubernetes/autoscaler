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

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	projectsSubstring = "/projects/"
	defaultDomainUrl  = "https://www.googleapis.com/compute/v1"
)

// ParseMigUrl expects url in format:
// https://www.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instanceGroups/<name>
func ParseMigUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(url, "instanceGroups")
}

// ParseIgmUrl expects url in format:
// https://www.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instanceGroupManagers/<name>
func ParseIgmUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(url, "instanceGroupManagers")
}

// ParseInstanceUrl expects url in format:
// https://www.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instances/<name>
func ParseInstanceUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(url, "instances")
}

// ParseInstanceUrlRef expects url in format:
// https://www.googleapis.com/compute/v1/projects/<project-id>/zones/<zone>/instances/<name>
// and returns a GceRef struct for it.
func ParseInstanceUrlRef(url string) (GceRef, error) {
	project, zone, name, err := parseGceUrl(url, "instances")
	if err != nil {
		return GceRef{}, err
	}
	return GceRef{
		Project: project,
		Zone:    zone,
		Name:    name,
	}, nil
}

// GenerateInstanceUrl generates url for instance.
func GenerateInstanceUrl(domainUrl string, ref GceRef) string {
	if domainUrl == "" {
		domainUrl = defaultDomainUrl
	}
	instanceUrlTemplate := domainUrl + projectsSubstring + "%s/zones/%s/instances/%s"
	return fmt.Sprintf(instanceUrlTemplate, ref.Project, ref.Zone, ref.Name)
}

// GenerateMigUrl generates url for instance.
func GenerateMigUrl(domainUrl string, ref GceRef) string {
	if domainUrl == "" {
		domainUrl = defaultDomainUrl
	}
	migUrlTemplate := domainUrl + projectsSubstring + "%s/zones/%s/instanceGroups/%s"
	return fmt.Sprintf(migUrlTemplate, ref.Project, ref.Zone, ref.Name)
}

// IsInstanceTemplateRegional determines whether or not an instance template is regional based on the url
func IsInstanceTemplateRegional(templateUrl string) (bool, error) {
	return regexp.MatchString("(/projects/.*[A-Za-z0-9]+.*/regions/)", templateUrl)
}

func parseGceUrl(url, expectedResource string) (project string, zone string, name string, err error) {
	reg := regexp.MustCompile(fmt.Sprintf("https://.*/projects/.*/zones/.*/%s/.*", expectedResource))
	errMsg := fmt.Errorf("wrong url: expected format <url>/projects/<project-id>/zones/<zone>/%s/<name>, got %s", expectedResource, url)
	if !reg.MatchString(url) {
		return "", "", "", errMsg
	}
	splitted := strings.Split(strings.Split(url, projectsSubstring)[1], "/")
	project = splitted[0]
	zone = splitted[2]
	name = splitted[4]
	return project, zone, name, nil
}
