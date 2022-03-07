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

package common

type sections struct {
	contains map[string]*section
}

func (ss sections) section(name string) *section {
	s, ok := ss.contains[name]
	if !ok {
		s = new(section)
		ss.contains[name] = s
	}
	return s
}

type section struct {
	content map[string]*value
}

func (s *section) key(name string) *value {
	v, ok := s.content[name]
	if !ok {
		v = new(value)
		s.content[name] = v
	}
	return v
}
