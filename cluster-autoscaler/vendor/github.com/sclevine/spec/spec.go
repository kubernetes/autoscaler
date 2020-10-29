package spec

import (
	"bytes"
	"io"
	"testing"
	"time"
)

// G defines a group of specs.
// Unlike other testing libraries, it is re-evaluated for each subspec.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
type G func(text string, f func(), opts ...Option)

// Pend skips all specs in the provided group.
//
// All Options are ignored.
func (g G) Pend(text string, f func(), _ ...Option) {
	g(text, f, func(c *config) { c.pend = true })
}

// Focus focuses the provided group.
// This skips all specs in the suite except the group and other focused specs.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
func (g G) Focus(text string, f func(), opts ...Option) {
	g(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// S defines a spec.
//
// Valid Options: Parallel
type S func(text string, f func(), opts ...Option)

// Before runs a function before each spec in the group.
func (s S) Before(f func()) {
	s("", f, func(c *config) { c.before = true })
}

// After runs a function after each spec in the group.
func (s S) After(f func()) {
	s("", f, func(c *config) { c.after = true })
}

// Pend skips the provided spec.
//
// All Options are ignored.
func (s S) Pend(text string, f func(), _ ...Option) {
	s(text, f, func(c *config) { c.pend = true })
}

// Focus focuses the provided spec.
// This skips all specs in the suite except the spec and other focused specs.
//
// Valid Options: Parallel
func (s S) Focus(text string, f func(), opts ...Option) {
	s(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// Out provides an dedicated writer for the test to store output.
// Reporters usually display the contents on test failure.
//
// Valid context: inside S blocks only, nil elsewhere
func (s S) Out() io.Writer {
	var out io.Writer
	s("", nil, func(c *config) {
		c.out = func(w io.Writer) {
			out = w
		}
	})
	return out
}

// Suite defines a top-level group of specs within a suite.
// Suite behaves like a top-level version of G.
// Unlike other testing libraries, it is re-evaluated for each subspec.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
type Suite func(text string, f func(*testing.T, G, S), opts ...Option) bool

// Before runs a function before each spec in the suite.
func (s Suite) Before(f func(*testing.T)) bool {
	return s("", func(t *testing.T, _ G, _ S) {
		t.Helper()
		f(t)
	}, func(c *config) { c.before = true })
}

// After runs a function after each spec in the suite.
func (s Suite) After(f func(*testing.T)) bool {
	return s("", func(t *testing.T, _ G, _ S) {
		t.Helper()
		f(t)
	}, func(c *config) { c.after = true })
}

// Pend skips the provided top-level group of specs.
//
// All Options are ignored.
func (s Suite) Pend(text string, f func(*testing.T, G, S), _ ...Option) bool {
	return s(text, f, func(c *config) { c.pend = true })
}

// Focus focuses the provided top-level group.
// This skips all specs in the suite except the group and other focused specs.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
func (s Suite) Focus(text string, f func(*testing.T, G, S), opts ...Option) bool {
	return s(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// Run executes the specs defined in each top-level group of the suite.
func (s Suite) Run(t *testing.T) bool {
	t.Helper()
	return s("", nil, func(c *config) { c.t = t })
}

// New creates an empty suite and returns an Suite function.
// The Suite function may be called to add top-level groups to the suite.
// The suite may be executed with Suite.Run.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
// Seed, Report
func New(text string, opts ...Option) Suite {
	var fs []func(*testing.T, G, S)
	return func(newText string, f func(*testing.T, G, S), newOpts ...Option) bool {
		cfg := options(newOpts).apply()
		if cfg.t == nil {
			fs = append(fs, func(t *testing.T, g G, s S) {
				var do func(string, func(), ...Option) = g
				if cfg.before || cfg.after {
					do = s
				}
				do(newText, func() { f(t, g, s) }, newOpts...)
			})
			return true
		}
		cfg.t.Helper()
		return Run(cfg.t, text, func(t *testing.T, g G, s S) {
			for _, f := range fs {
				f(t, g, s)
			}
		}, opts...)
	}
}

// Run immediately executes the provided specs as a suite.
// Unlike other testing libraries, it is re-evaluated for each spec.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
// Seed, Report
func Run(t *testing.T, text string, f func(*testing.T, G, S), opts ...Option) bool {
	t.Helper()
	cfg := options(opts).apply()
	n := &node{
		text:  []string{text},
		seed:  defaultZero64(cfg.seed, time.Now().Unix()),
		order: cfg.order.or(orderSequential),
		scope: cfg.scope.or(scopeLocal),
		nest:  cfg.nest.or(nestOff),
		pend:  cfg.pend,
		focus: cfg.focus,
	}
	report := cfg.report
	plan := n.parse(f)

	var specs chan Spec
	if report != nil {
		report.Start(t, plan)
		specs = make(chan Spec, plan.Total)
		done := make(chan struct{})
		defer func() {
			close(specs)
			<-done
		}()
		go func() {
			report.Specs(t, specs)
			close(done)
		}()
	}

	return n.run(t, func(t *testing.T, n node) {
		t.Helper()
		buffer := &bytes.Buffer{}
		defer func() {
			if specs == nil {
				return
			}
			specs <- Spec{
				Text:     n.text,
				Failed:   t.Failed(),
				Skipped:  t.Skipped(),
				Focused:  n.focus,
				Parallel: n.order == orderParallel,
				Out:      buffer,
			}
		}()
		switch {
		case n.pend, plan.HasFocus && !n.focus:
			t.SkipNow()
		case n.order == orderParallel:
			t.Parallel()
		}

		var spec, group func()
		hooks := newHooks()
		group = func() {}

		f(t, func(_ string, f func(), _ ...Option) {
			switch {
			case len(n.loc) == 1, n.loc[0] > 0:
				n.loc[0]--
			case n.loc[0] == 0:
				group = func() {
					n.loc = n.loc[1:]
					hooks.next()
					group = func() {}
					f()
					group()
				}
				n.loc[0]--
			}
		}, func(_ string, f func(), opts ...Option) {
			cfg := options(opts).apply()
			switch {
			case cfg.out != nil:
				cfg.out(buffer)
			case cfg.before:
				hooks.before(f)
			case cfg.after:
				hooks.after(f)
			case spec != nil:
			case len(n.loc) > 1, n.loc[0] > 0:
				n.loc[0]--
			default:
				spec = f
			}
		})
		group()

		if spec == nil {
			t.Fatal("Failed to locate spec.")
		}
		hooks.run(t, spec)
	})
}

type specHooks struct {
	first, last *specHook
}

type specHook struct {
	before, after []func()
	next          *specHook
}

func newHooks() specHooks {
	h := &specHook{}
	return specHooks{first: h, last: h}
}

func (s specHooks) run(t *testing.T, spec func()) {
	t.Helper()
	for h := s.first; h != nil; h = h.next {
		defer run(t, h.after...)
		run(t, h.before...)
	}
	run(t, spec)
}

func (s specHooks) before(f func()) {
	s.last.before = append(s.last.before, f)
}

func (s specHooks) after(f func()) {
	s.last.after = append(s.last.after, f)
}

func (s *specHooks) next() {
	s.last.next = &specHook{}
	s.last = s.last.next
}

func run(t *testing.T, fs ...func()) {
	t.Helper()
	for _, f := range fs {
		f()
	}
}

// Pend skips all specs in the top-level group.
//
// All Options are ignored.
func Pend(t *testing.T, text string, f func(*testing.T, G, S), _ ...Option) bool {
	t.Helper()
	return Run(t, text, f, func(c *config) { c.pend = true })
}

// Focus focuses every spec in the provided suite.
// This is useful as a shortcut for unfocusing all focused specs.
//
// Valid Options:
// Sequential, Random, Reverse, Parallel
// Local, Global, Flat, Nested
// Seed, Report
func Focus(t *testing.T, text string, f func(*testing.T, G, S), opts ...Option) bool {
	t.Helper()
	return Run(t, text, f, append(opts, func(c *config) { c.focus = true })...)
}

// A Plan provides a Reporter with information about a suite.
type Plan struct {
	Text      string
	Total     int
	Pending   int
	Focused   int
	Seed      int64
	HasRandom bool
	HasFocus  bool
}

// A Spec provides a Reporter with information about a spec immediately after
// the spec completes.
type Spec struct {
	Text     []string
	Failed   bool
	Skipped  bool
	Focused  bool
	Parallel bool
	Out      io.Reader
}

// A Reporter is provided with information about a suite as it runs.
type Reporter interface {

	// Start provides the Reporter with a Plan that describes the suite.
	// No specs will run until the Start method call finishes.
	Start(*testing.T, Plan)

	// Specs provides the Reporter with a channel of Specs.
	// The specs will start running concurrently with the Specs method call.
	// The Run method will not complete until the Specs method call completes.
	Specs(*testing.T, <-chan Spec)
}
