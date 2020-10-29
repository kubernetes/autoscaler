package spec

import (
	"io"
	"testing"
)

// An Option controls the behavior of a suite, group, or spec.
// Options are inherited by subgroups and subspecs.
//
// Example:
// If the top-level Run group is specified as Random(), each subgroup will
// inherit the Random() order. This means that each group will be randomized
// individually, unless another ordering is specified on any of the subgroups.
// If the Run group is also passed Global(), then all specs inside Run will run
// in completely random order, regardless of any ordering specified on the
// subgroups.
type Option func(*config)

// Report specifies a Reporter for a suite.
//
// Valid Option for:
// New, Run, Focus, Pend
func Report(r Reporter) Option {
	return func(c *config) {
		c.report = r
	}
}

// Seed specifies the random seed used for any randomized specs in a Run block.
// The random seed is always displayed before specs are run.
// If not specified, the current time is used.
//
// Valid Option for:
// New, Run, Focus, Pend
func Seed(s int64) Option {
	return func(c *config) {
		c.seed = s
	}
}

// Sequential indicates that a group of specs should be run in order.
// This is the default behavior.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Sequential() Option {
	return func(c *config) {
		c.order = orderSequential
	}
}

// Random indicates that a group of specs should be run in random order.
// Randomization is per group, such that all groupings are maintained.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Random() Option {
	return func(c *config) {
		c.order = orderRandom
	}
}

// Reverse indicates that a group of specs should be run in reverse order.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Reverse() Option {
	return func(c *config) {
		c.order = orderReverse
	}
}

// Parallel indicates that a spec or group of specs should be run in parallel.
// This Option is equivalent to t.Parallel().
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend, S
func Parallel() Option {
	return func(c *config) {
		c.order = orderParallel
	}
}

// Local indicates that the test order applies to each subgroup individually.
// For example, a group with Random() and Local() will run all subgroups and
// specs in random order, and each subgroup will be randomized, but specs in
// different subgroups will not be interleaved.
// This is the default behavior.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Local() Option {
	return func(c *config) {
		c.scope = scopeLocal
	}
}

// Global indicates that test order applies globally to all descendant specs.
// For example, a group with Random() and Global() will run all descendant
// specs in random order, regardless of subgroup. Specs in different subgroups
// may be interleaved.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Global() Option {
	return func(c *config) {
		c.scope = scopeGlobal
	}
}

// Flat indicates that a parent subtest should not be created for the group.
// This is the default behavior.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Flat() Option {
	return func(c *config) {
		c.nest = nestOff
	}
}

// Nested indicates that a parent subtest should be created for the group.
// This allows for more control over parallelism.
//
// Valid Option for:
// New, Run, Focus, Pend, Suite, Suite.Focus, Suite.Pend, G, G.Focus, G.Pend
func Nested() Option {
	return func(c *config) {
		c.nest = nestOn
	}
}

type order int

const (
	orderInherit order = iota
	orderSequential
	orderParallel
	orderRandom
	orderReverse
)

func (o order) or(last order) order {
	return order(defaultZero(int(o), int(last)))
}

type scope int

const (
	scopeInherit scope = iota
	scopeLocal
	scopeGlobal
)

func (s scope) or(last scope) scope {
	return scope(defaultZero(int(s), int(last)))
}

type nest int

const (
	nestInherit nest = iota
	nestOff
	nestOn
)

func (n nest) or(last nest) nest {
	return nest(defaultZero(int(n), int(last)))
}

func defaultZero(next, last int) int {
	if next == 0 {
		return last
	}
	return next
}

func defaultZero64(next, last int64) int64 {
	if next == 0 {
		return last
	}
	return next
}

type config struct {
	seed   int64
	order  order
	scope  scope
	nest   nest
	pend   bool
	focus  bool
	before bool
	after  bool
	t      *testing.T
	out    func(io.Writer)
	report Reporter
}

type options []Option

func (o options) apply() *config {
	cfg := &config{}
	for _, opt := range o {
		opt(cfg)
	}
	return cfg
}
