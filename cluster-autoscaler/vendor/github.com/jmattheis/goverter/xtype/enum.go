package xtype

import (
	"go/types"
	"sort"

	"github.com/jmattheis/goverter/enum"
)

type Enum struct {
	enum.Enum
	OK bool
}

func (t *Type) Enum(cfg *enum.Config) *Enum {
	if !t.Named {
		return disabled
	}

	if t.enum == nil {
		t.enum = loadEnum(t.NamedType, cfg)
	}
	return t.enum
}

func loadEnum(t *types.Named, cfg *enum.Config) *Enum {
	path := t.Obj().Pkg().Path()
	name := t.Obj().Name()

	if !cfg.Enabled || cfg.Excludes.Matches(path, name) {
		return disabled
	}

	e, ok := enum.Detect(t)
	return &Enum{OK: ok, Enum: e}
}

func (e Enum) SortedMembers() []string {
	var m []string
	for member := range e.Members {
		m = append(m, member)
	}
	sort.Strings(m)
	return m
}

var disabled = &Enum{OK: false}
