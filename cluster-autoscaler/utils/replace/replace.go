/*
Copyright 2021 The Kubernetes Authors.

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

package replace

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
)

// Replacement contains a regular expresssion and a replacement string.
type Replacement struct {
	Re          regexp.Regexp
	Replacement string
}

func (r Replacement) String() string {
	// This is intentionally simplistic: we don't bother checking
	// whether ";" occurs inside the regular expression or the
	// replacement string.
	return ";" + r.Re.String() + ";" + r.Replacement + ";"
}

// Replacements contains multiple regular expressions and their replacements.
type Replacements []Replacement

var _ flag.Value = &Replacements{}

// ApplyToPair turns a key/value pair into a single <key>=<value> string,
// applies all regular expressions, then splits again. Key and value
// may become empty.
func (r *Replacements) ApplyToPair(k, v string) (string, string) {
	if r == nil {
		return k, v
	}
	str := k + "=" + v
	for _, repl := range *r {
		str = repl.Re.ReplaceAllString(str, repl.Replacement)
	}
	parts := strings.SplitN(str, "=", 2)
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func (r *Replacements) String() string {
	var parts []string
	for _, repl := range *r {
		parts = append(parts, repl.String())
	}
	return strings.Join(parts, " ")
}

const format = "must be of the form <sep><regexp><sep><replacement><sep>"

// Set adds one replacement of the form <sep><regexp><sep><replacement><sep>.
func (r *Replacements) Set(str string) error {
	if len(str) < 3 {
		return errors.New(format + ": too short")
	}
	sep := str[0]
	if str[len(str)-1] != sep {
		return errors.New(format + ": separator at start and end does not match")
	}
	str = str[1 : len(str)-1]
	parts := strings.Split(str, string(sep))
	if len(parts) != 2 {
		return errors.New(format + ": need exactly one separator between regular expression and replacement")
	}
	re, err := regexp.Compile(parts[0])
	if err != nil {
		return fmt.Errorf("%s: regular expression invalid: %v", format, err)
	}
	*r = append(*r, Replacement{Re: *re, Replacement: parts[1]})
	return nil
}
