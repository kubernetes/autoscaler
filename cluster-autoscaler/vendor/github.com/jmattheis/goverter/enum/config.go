package enum

import "regexp"

type Config struct {
	Unknown  string
	Enabled  bool
	Excludes IDPatterns
}

type IDPattern struct {
	Path *regexp.Regexp
	Name *regexp.Regexp
}

type IDPatterns []IDPattern

func (ids IDPatterns) Matches(path, name string) bool {
	for _, id := range ids {
		if id.Path.MatchString(path) && id.Name.MatchString(name) {
			return true
		}
	}
	return false
}
