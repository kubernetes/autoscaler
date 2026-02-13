/*
Copyright 2019 The Kubernetes Authors.

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

package verdacloud

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var shellAssignmentRe = regexp.MustCompile(`^\s*(export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*?)(\s+#.*)?$`)

type shellQuoteStyle int

const (
	shellQuoteNone shellQuoteStyle = iota
	shellQuoteDouble
	shellQuoteSingle
)

// injectEnvVarsIntoScript rewrites shell variable assignments with values from envMap.
// Skips heredoc blocks to avoid corrupting embedded documents.
func injectEnvVarsIntoScript(script []byte, envMap map[string]string) []byte {
	var out bytes.Buffer
	sc := bufio.NewScanner(bytes.NewReader(script))
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	inHeredoc := false
	heredocEnd := ""

	for sc.Scan() {
		line := sc.Text()

		if !inHeredoc {
			if del, ok := parseHeredocDelimiter(line); ok {
				inHeredoc = true
				heredocEnd = del
				out.WriteString(line)
				out.WriteByte('\n')
				continue
			}
			out.WriteString(rewriteShellAssignment(line, envMap))
			out.WriteByte('\n')
			continue
		}

		if strings.TrimRight(line, "\r\n") == heredocEnd {
			inHeredoc = false
			heredocEnd = ""
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.Bytes()
}

// parseHeredocDelimiter detects heredoc syntax (<<EOF or <<'EOF') and returns the delimiter.
func parseHeredocDelimiter(line string) (string, bool) {
	hd := regexp.MustCompile(`<<-?'?([A-Za-z0-9_]+)'?`)
	m := hd.FindStringSubmatch(line)
	if len(m) == 2 {
		return m[1], true
	}
	return "", false
}

// rewriteShellAssignment replaces the value of a VAR=value line if VAR is in envMap.
func rewriteShellAssignment(line string, envMap map[string]string) string {
	m := shellAssignmentRe.FindStringSubmatch(line)
	if m == nil {
		return line
	}
	exportPrefix := m[1]
	varName := m[2]
	rawVal := strings.TrimSpace(m[3])
	comment := ""
	if len(m) >= 5 && m[4] != "" {
		comment = m[4]
	}

	newVal, ok := envMap[varName]
	if !ok {
		return line
	}

	quoted := detectShellQuoteStyle(rawVal)
	repl := formatWithShellQuotes(newVal, quoted)

	// Keep bare if safe, quote otherwise
	if quoted == shellQuoteNone && isShellSafeUnquoted(newVal) {
		repl = newVal
	}

	var b strings.Builder
	if exportPrefix != "" {
		b.WriteString(strings.TrimRight(exportPrefix, " "))
		b.WriteByte(' ')
	}
	b.WriteString(varName)
	b.WriteByte('=')
	b.WriteString(repl)
	if comment != "" {
		b.WriteString(comment)
	}
	return b.String()
}

// detectShellQuoteStyle determines if a value uses double quotes, single quotes, or none.
func detectShellQuoteStyle(s string) shellQuoteStyle {
	s = strings.TrimSpace(s)
	if s == "" {
		return shellQuoteNone
	}
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(trimCarriageReturn(s), `"`) && len(s) >= 2 {
		return shellQuoteDouble
	}
	if strings.HasPrefix(s, `'`) && strings.HasSuffix(trimCarriageReturn(s), `'`) && len(s) >= 2 {
		return shellQuoteSingle
	}
	return shellQuoteNone
}

// trimCarriageReturn removes trailing \r characters (for Windows line endings).
func trimCarriageReturn(s string) string {
	return strings.TrimRight(s, "\r")
}

// isShellSafeUnquoted returns true if s can be used unquoted in shell (alphanumeric + safe chars).
func isShellSafeUnquoted(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '_', '.', '/', ':', '-', '@':
			continue
		default:
			return false
		}
	}
	return true
}

// escapeForDoubleQuotes escapes \, ", and $ for use inside double-quoted shell strings.
func escapeForDoubleQuotes(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '"', '$':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// formatWithShellQuotes wraps value with appropriate shell quoting based on style.
func formatWithShellQuotes(value string, style shellQuoteStyle) string {
	switch style {
	case shellQuoteSingle:
		if strings.ContainsRune(value, '\'') {
			return `"` + escapeForDoubleQuotes(value) + `"`
		}
		return `'` + value + `'`
	case shellQuoteDouble:
		return `"` + escapeForDoubleQuotes(value) + `"`
	default:
		return `"` + escapeForDoubleQuotes(value) + `"`
	}
}

func convertConfigLabelsToK8sLabels(labels []string, asg *Asg) string {
	if asg == nil {
		return ""
	}
	if len(labels) == 0 {
		labels = make([]string, 2)
	}

	labels = append(labels, fmt.Sprintf("%s=%s", NodeGroupLabelKey, asg.Name))

	if isGPUInstanceType(asg.instanceType) {
		labels = append(labels, fmt.Sprintf("%s=%s", AcceleratorLabel, asg.instanceType))
	}

	_k8sLabels := strings.Join(labels, ",")
	return _k8sLabels
}

func parseAsgSpec(spec string) (*VerdacloudAsgSpec, error) {
	parts := strings.Split(spec, ":")
	if len(parts) != 4 && len(parts) != 5 {
		return nil, fmt.Errorf("invalid ASG spec (expected min:max:instance-type:asg-name[:hostname-prefix]): %s", spec)
	}

	minSize, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid min size: %s", parts[0])
	}
	maxSize, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid max size: %s", parts[1])
	}

	validName := regexp.MustCompile(`^[a-z0-9A-Z]+[a-z0-9A-Z\-\.\_]*[a-z0-9A-Z]+$|^[a-z0-9A-Z]{1}$`)
	asgName := parts[3]
	if !validName.MatchString(asgName) {
		return nil, fmt.Errorf("invalid ASG name: %s", asgName)
	}

	hostnamePrefix := ""
	if len(parts) == 5 && parts[4] != "" {
		hostnamePrefix = parts[4]
		if !validName.MatchString(hostnamePrefix) {
			return nil, fmt.Errorf("invalid hostname prefix: %s", hostnamePrefix)
		}
	}

	return &VerdacloudAsgSpec{
		minSize:        minSize,
		maxSize:        maxSize,
		instanceType:   parts[2],
		name:           asgName,
		hostnamePrefix: hostnamePrefix,
	}, nil
}

func isGPUInstanceType(instanceType string) bool {
	return !strings.HasPrefix(strings.ToUpper(instanceType), "CPU.")
}

func instanceRefFromProviderId(providerId string) (*InstanceRef, error) {
	providerIdBase := strings.TrimPrefix(providerId, verdacloudProviderIDPrefix)
	parts := strings.Split(providerIdBase, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid provider ID: %s", providerId)
	}
	return &InstanceRef{Hostname: parts[len(parts)-1], ProviderID: providerId}, nil
}

func extractAsgNameFromHostname(hostname string) (string, error) {
	separator := fmt.Sprintf("-%s-", ASG_SEPARATOR_MAGIC_NUMBER)

	parts := strings.Split(hostname, separator)
	if len(parts) == 2 {
		asgName := parts[0]
		if asgName == "" {
			return "", fmt.Errorf("empty ASG name extracted from hostname: %s", hostname)
		}
		return asgName, nil
	}

	return "", fmt.Errorf("hostname does not contain magic separator '%s': %s", separator, hostname)
}
