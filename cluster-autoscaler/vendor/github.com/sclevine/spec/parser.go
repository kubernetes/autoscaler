package spec

import (
	"math/rand"
	"strings"
	"testing"
)

type node struct {
	text  []string
	loc   []int
	seed  int64
	order order
	scope scope
	nest  nest
	pend  bool
	focus bool
	nodes tree
}

func (n *node) parse(f func(*testing.T, G, S)) Plan {
	// TODO: validate Options
	plan := Plan{
		Text: strings.Join(n.text, "/"),
		Seed: n.seed,
	}
	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		n.add(text, cfg, tree{})
		parent := n
		n = n.last()
		plan.update(n)
		defer func() {
			n.level()
			n.sort()
			n = parent
		}()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after || cfg.out != nil {
			return
		}
		n.add(text, cfg, nil)
		plan.update(n.last())
	})
	n.level()
	n.sort()
	return plan
}

func (p *Plan) update(n *node) {
	if n.focus && !n.pend {
		p.HasFocus = true
	}
	if n.order == orderRandom {
		p.HasRandom = true
	}
	if n.nodes == nil {
		p.Total++
		if n.focus && !n.pend {
			p.Focused++
		} else if n.pend {
			p.Pending++
		}
	}
}

func (n *node) add(text string, cfg *config, nodes tree) {
	name := n.text
	if n.nested() {
		name = nil
	}
	n.nodes = append(n.nodes, node{
		text:  append(append([]string(nil), name...), text),
		loc:   append(append([]int(nil), n.loc...), len(n.nodes)),
		seed:  n.seed,
		order: cfg.order.or(n.order),
		scope: cfg.scope.or(n.scope),
		nest:  cfg.nest.or(n.nest),
		pend:  cfg.pend || n.pend,
		focus: cfg.focus || n.focus,
		nodes: nodes,
	})
}

func (n *node) sort() {
	nodes := n.nodes
	switch n.order {
	case orderRandom:
		r := rand.New(rand.NewSource(n.seed))
		for i := len(nodes) - 1; i > 0; i-- {
			j := r.Intn(i + 1)
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	case orderReverse:
		last := len(nodes) - 1
		for i := 0; i < len(nodes)/2; i++ {
			nodes[i], nodes[last-i] = nodes[last-i], nodes[i]
		}
	}
}

func (n *node) level() {
	nodes := n.nodes
	switch n.scope {
	case scopeGlobal:
		var flat tree
		for _, child := range nodes {
			if child.nodes == nil || child.scope == scopeLocal {
				flat = append(flat, child)
			} else {
				flat = append(flat, child.nodes...)
			}
		}
		n.nodes = flat
	}
}

func (n *node) last() *node {
	return &n.nodes[len(n.nodes)-1]
}

func (n *node) nested() bool {
	return n.nest == nestOn || len(n.loc) == 0
}

func (n node) run(t *testing.T, f func(*testing.T, node)) bool {
	t.Helper()
	name := strings.Join(n.text, "/")
	switch {
	case n.nodes == nil:
		return t.Run(name, func(t *testing.T) { f(t, n) })
	case n.nested():
		return t.Run(name, func(t *testing.T) { n.nodes.run(t, f) })
	default:
		return n.nodes.run(t, f)
	}
}

type tree []node

func (ns tree) run(t *testing.T, f func(*testing.T, node)) bool {
	t.Helper()
	ok := true
	for _, n := range ns {
		ok = n.run(t, f) && ok
	}
	return ok
}
