/*
Copyright 2023 The Kubernetes Authors.

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

package ini

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

func isNewline(b []rune) bool {
	if len(b) == 0 {
		return false
	}

	if b[0] == '\n' {
		return true
	}

	if len(b) < 2 {
		return false
	}

	return b[0] == '\r' && b[1] == '\n'
}

func newNewlineToken(b []rune) (Token, int, error) {
	i := 1
	if b[0] == '\r' && isNewline(b[1:]) {
		i++
	}

	if !isNewline([]rune(b[:i])) {
		return emptyToken, 0, NewParseError("invalid new line token")
	}

	return newToken(TokenNL, b[:i], NoneType), i, nil
}
