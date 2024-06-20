package builder

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

// Path defines the path inside an error message.
type Path struct {
	Prefix     string
	SourceID   string
	TargetID   string
	SourceType string
	TargetType string
}

// Error defines a conversion error.
type Error struct {
	Path  []*Path
	Cause string
}

// NewError creates an error.
func NewError(cause string) *Error {
	return &Error{Cause: cause, Path: []*Path{}}
}

// Lift appends the path to the error.
func (e *Error) Lift(paths ...*Path) *Error {
	e.Path = append(paths, e.Path...)
	return e
}

// ToString converts the error into a string.
func ToString(err *Error) string {
	if len(err.Path) == 0 {
		panic("oops that shouldn't happen")
	}

	sourcePaths := 0
	targetPaths := 0
	for _, path := range err.Path {
		if path.SourceType != "" {
			sourcePaths++
		}
		if path.TargetType != "" {
			targetPaths++
		}
	}

	end := 2 + (sourcePaths+targetPaths)*2 - 1
	sourceLine := (sourcePaths * 2)
	targetLine := sourceLine + 1

	lines := make([]string, end+1)

	sourceTypeLine := 0
	targetTypeLine := end
	for i := 0; i < len(err.Path); i++ {
		path := err.Path[i]
		padding := int(math.Max(float64(len(path.SourceID)), float64(len(path.TargetID))))

		if path.SourceType != "" {
			lines[sourceTypeLine] += space(len(path.Prefix)) + "| " + path.SourceType

			for j := sourceTypeLine + 1; j < sourceLine; j++ {
				lines[j] += space(len(path.Prefix)) + "|" + space(padding-1)
			}
			sourceTypeLine += 2
		} else {
			for j := sourceTypeLine; j < sourceLine; j++ {
				lines[j] += space(len(path.Prefix) + padding)
			}
		}

		lines[sourceLine] += path.Prefix + path.SourceID + space(padding-len(path.SourceID))

		if path.TargetType != "" {
			lines[targetLine] += path.Prefix + path.TargetID + space(padding-len(path.TargetID))

			for j := targetTypeLine - 1; j > targetLine; j-- {
				lines[j] += space(len(path.Prefix)) + "|" + space(padding-1)
			}
			lines[targetTypeLine] += space(len(path.Prefix)) + "| " + path.TargetType
			targetTypeLine -= 2
		} else {
			for j := targetTypeLine; j >= targetLine; j-- {
				lines[j] += space(len(path.Prefix) + padding)
			}
		}
	}

	buf := bytes.Buffer{}
	for _, line := range lines {
		_, _ = fmt.Fprintln(&buf, strings.TrimSpace(line))
	}
	fmt.Fprintln(&buf)
	fmt.Fprint(&buf, err.Cause)
	return buf.String()
}

func space(l int) string {
	return strings.Repeat(" ", l)
}
