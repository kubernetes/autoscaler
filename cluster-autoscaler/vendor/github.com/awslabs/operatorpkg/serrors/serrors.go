package serrors

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
	"go.uber.org/multierr"
)

// Error is a structured error that stores structured errors and values alongside the error
type Error struct {
	error
	keysAndValues map[string]any
}

// Unwrap returns the unwrapped error
func (e *Error) Unwrap() error {
	return e.error
}

// Error returns the string representation of the error
func (e *Error) Error() string {
	var elems []string
	keys := lo.Keys(e.keysAndValues)
	sort.Strings(keys) // sort keys so we always get a consistent ordering
	for _, k := range keys {
		v := e.keysAndValues[k]
		elems = append(elems, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("%s (%s)", e.error.Error(), strings.Join(elems, ", "))
}

// WithValues injects additional structured keys and values into the error
func (e *Error) WithValues(keysAndValues ...any) *Error {
	lo.Must0(validateKeysAndValues(keysAndValues))
	if e.keysAndValues == nil {
		e.keysAndValues = map[string]any{}
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		e.keysAndValues[keysAndValues[i].(string)] = keysAndValues[i+1]
	}
	return e
}

// Wrap wraps and existing error with additional structured keys and values
func Wrap(err error, keysAndValues ...any) error {
	e := &Error{error: err}
	return e.WithValues(keysAndValues...)
}

func validateKeysAndValues(keysAndValues []any) error {
	if len(keysAndValues)%2 != 0 {
		return fmt.Errorf("keysAndValues must have an even number of elements")
	}
	for i := 0; i < len(keysAndValues); i += 2 {
		if _, ok := keysAndValues[i].(string); !ok {
			return fmt.Errorf("keys must be strings")
		}
	}
	return nil
}

// UnwrapValues returns a combined set of keys and values from every wrapped error
func UnwrapValues(err error) (res []any) {
	values := map[string][]any{}
	for err != nil {
		for _, elem := range multierr.Errors(err) {
			if e, ok := elem.(*Error); ok {
				for k, v := range e.keysAndValues {
					if _, mOk := values[k]; mOk {
						values[k] = append(values[k], v)
					} else {
						values[k] = []any{v}
					}
				}
			}
		}
		err = errors.Unwrap(err)
	}
	if len(values) == 0 {
		return nil
	}
	keys := lo.Keys(values)
	sort.Strings(keys) // sort keys so we always get a consistent ordering
	for _, k := range keys {
		v := values[k]
		if len(v) == 1 {
			res = append(res, k, v[0])
		} else {
			res = append(res, fmt.Sprintf("%ss", k), v)
		}
	}
	return res
}
