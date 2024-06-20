package namer

import (
	"fmt"

	"github.com/jmattheis/goverter/xtype"
)

// New returns a new namer.
func New() *Namer {
	return &Namer{lookup: map[string]struct{}{xtype.ThisVar: {}}}
}

// Namer keeps track of used variable names.
type Namer struct {
	lookup map[string]struct{}
	First  string
}

var indexVars = []string{"i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

// Index returns an usused index variable name.
func (m *Namer) Index() string {
	for i := 1; ; i++ {
		for _, v := range indexVars {
			name := v
			if i > 1 {
				name += fmt.Sprint(i)
			}
			if m.Register(name) {
				return name
			}
		}
	}
}

// Map returns an usused key and value variable name.
func (m *Namer) Map() (string, string) {
	for i := 0; ; i++ {
		key := "key"
		value := "value"
		if i > 1 {
			key += fmt.Sprint(i)
			value += fmt.Sprint(i)
		}
		_, okKey := m.lookup[key]
		_, okValue := m.lookup[value]
		if !okKey && !okValue {
			m.lookup[key] = struct{}{}
			m.lookup[value] = struct{}{}
			return key, value
		}
	}
}

// Register registers a variable as used.
func (m *Namer) Register(name string) bool {
	if _, ok := m.lookup[name]; !ok {
		if m.First == "" {
			m.First = name
		}
		m.lookup[name] = struct{}{}
		return true
	}
	return false
}

// Name returns an unused variable name that contains the passed name.
func (m *Namer) Name(name string) string {
	for i := 1; ; i++ {
		numberedName := name
		if i > 1 {
			numberedName += fmt.Sprint(i)
		}
		if m.Register(numberedName) {
			return numberedName
		}
	}
}
