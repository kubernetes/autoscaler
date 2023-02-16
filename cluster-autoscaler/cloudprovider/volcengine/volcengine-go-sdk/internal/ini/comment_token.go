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

// isComment will return whether or not the next byte(s) is a
// comment.
func isComment(b []rune) bool {
	if len(b) == 0 {
		return false
	}

	switch b[0] {
	case ';':
		return true
	case '#':
		return true
	}

	return false
}

// newCommentToken will create a comment token and
// return how many bytes were read.
func newCommentToken(b []rune) (Token, int, error) {
	i := 0
	for ; i < len(b); i++ {
		if b[i] == '\n' {
			break
		}

		if len(b)-i > 2 && b[i] == '\r' && b[i+1] == '\n' {
			break
		}
	}

	return newToken(TokenComment, b[:i], NoneType), i, nil
}
