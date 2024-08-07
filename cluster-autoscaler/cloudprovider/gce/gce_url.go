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
	"net/url"
	"path"
	"regexp"
)

const (
	projectsSubstring  = "/projects/"
	defaultDomainUrl   = "https://www.googleapis.com/compute/v1"
	anyHttpsUrlPattern = "https://.*/"
)

// ParseMigUrl expects url in format:
// https://.*/projects/<project-id>/zones/<zone>/instanceGroups/<name>
func ParseMigUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(anyHttpsUrlPattern, url, "instanceGroups")
}

// ParseIgmUrl expects url in format:
// https://.*/<project-id>/zones/<zone>/instanceGroupManagers/<name>
func ParseIgmUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(anyHttpsUrlPattern, url, "instanceGroupManagers")
}

// ParseIgmUrlRef expects url in format:
// projects/<project-id>/zones/<zone>/instanceGroupManagers/<name>
// and returns a GceRef struct for it.
func ParseIgmUrlRef(url string) (GceRef, error) {
	project, zone, name, err := parseGceUrl("", url, "instanceGroupManagers")
	if err != nil {
		return GceRef{}, err
	}
	return GceRef{
		Project: project,
		Zone:    zone,
		Name:    name,
	}, nil
}

// ParseInstanceUrl expects url in format:
// https://.*/<project-id>/zones/<zone>/instances/<name>
func ParseInstanceUrl(url string) (project string, zone string, name string, err error) {
	return parseGceUrl(anyHttpsUrlPattern, url, "instances")
}

// ParseInstanceUrlRef expects url in format:
// https://.*/projects/<project-id>/zones/<zone>/instances/<name>
// and returns a GceRef struct for it.
func ParseInstanceUrlRef(url string) (GceRef, error) {
	project, zone, name, err := parseGceUrl(anyHttpsUrlPattern, url, "instances")
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

// InstanceTemplateNameFromUrl retrieves name of the Instance Template from the url.
func InstanceTemplateNameFromUrl(instanceTemplateLink string) (InstanceTemplateName, error) {
	templateUrl, err := url.Parse(instanceTemplateLink)
	if err != nil {
		return InstanceTemplateName{}, err
	}
	regional, err := IsInstanceTemplateRegional(templateUrl.String())
	if err != nil {
		return InstanceTemplateName{}, err
	}

	_, templateName := path.Split(templateUrl.EscapedPath())
	return InstanceTemplateName{templateName, regional}, nil
}

func parseGceUrl(prefix, url, expectedResource string) (project string, zone string, name string, err error) {
	reg := regexp.MustCompile(fmt.Sprintf("%sprojects/.*/zones/.*/%s/.*", prefix, expectedResource))
	errMsg := fmt.Errorf("wrong url: expected format %sprojects/<project-id>/zones/<zone>/%s/<name>, got %s", prefix, expectedResource, url)
	if !reg.MatchString(url) {
		return "", "", "", errMsg
	}

	subMatches := regexp.MustCompile(fmt.Sprintf("%sprojects/(.*)/zones/(.*)/%s/(.*)", prefix, expectedResource)).FindStringSubmatch(url)
	project = subMatches[1]
	zone = subMatches[2]
	name = subMatches[3]
	return project, zone, name, nil
}
