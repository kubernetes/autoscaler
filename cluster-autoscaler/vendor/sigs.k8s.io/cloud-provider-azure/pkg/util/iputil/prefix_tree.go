/*
Copyright 2024 The Kubernetes Authors.

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

package iputil

import (
	"net/netip"
)

type prefixTreeNode struct {
	masked bool
	prefix netip.Prefix

	p *prefixTreeNode // parent node
	l *prefixTreeNode // left child node
	r *prefixTreeNode // right child node
}

// pruneToRoot  checks if the current node and its sibling are masked,
// and if so, marks their parent as masked and removes both children.
// This process is repeated up the tree until a node with an unmasked sibling is found.
//
// The process can be visualized as follows:
//
//	Before:           After:
//	   P                 P (masked)
//	  / \               / \
//	 A   B     ->      X   X
//	(M) (M)
//
// Where:
//
//	P: Parent node
//	A, B: Child nodes
//	M: Masked
//	X: Removed
//
// This method helps to optimize the tree structure by condensing fully masked subtrees.
func (n *prefixTreeNode) pruneToRoot() {
	var node = n
	for node.p != nil {
		p := node.p
		if p.l == nil || !p.l.masked {
			break
		}
		if p.r == nil || !p.r.masked {
			break
		}
		p.masked = true
		p.l, p.r = nil, nil
		node = p
	}
}

// prefixTree represents a tree structure for storing and managing IP prefixes.
// It efficiently handles prefix aggregation, merging of overlapping prefixes,
// and collapsing of neighboring prefixes.
//
// The tree is structured as follows:
// - Each node represents a bit in the IP address
// - Left child represents a 0 bit, right child represents a 1 bit
// - Masked nodes indicate the end of a prefix
// - Unused branches are represented by nil pointers
//
// Example tree for 128.0.0.0/4 (binary 1000 0000):
//
//	    0 (0.0.0.0/0)
//	   / \
//	  X   1 (128.0.0.0/1)
//	     / \
//	    0   X
//	   / \
//	  0   X
//	 / \
//	0*  X
//
// Where:
// * denotes a masked node (prefix end)
// X denotes an unused branch (nil pointer)
type prefixTree struct {
	maxBits int
	root    *prefixTreeNode
}

func newPrefixTreeForIPv4() *prefixTree {
	return &prefixTree{
		maxBits: 32,
		root: &prefixTreeNode{
			prefix: netip.MustParsePrefix("0.0.0.0/0"),
		},
	}
}

func newPrefixTreeForIPv6() *prefixTree {
	return &prefixTree{
		maxBits: 128,
		root: &prefixTreeNode{
			prefix: netip.MustParsePrefix("::/0"),
		},
	}
}

// Add adds a prefix to the tree.
func (t *prefixTree) Add(prefix netip.Prefix) {
	var (
		n    = t.root
		bits = prefix.Addr().AsSlice()
	)
	for i := 0; i < prefix.Bits(); i++ {
		if n.masked {
			break // It's already masked, the rest of the bits are irrelevant
		}

		var bit = bits[i/8] >> (7 - i%8) & 1
		switch bit {
		case 0:
			if n.l == nil {
				next, err := prefix.Addr().Prefix(i + 1)
				if err != nil {
					panic("unreachable: invalid prefix")
				}
				n.l = &prefixTreeNode{
					prefix: next,
					p:      n,
				}
			}
			n = n.l
		case 1:
			if n.r == nil {
				next, err := prefix.Addr().Prefix(i + 1)
				if err != nil {
					panic("unreachable: invalid prefix")
				}
				n.r = &prefixTreeNode{
					prefix: next,
					p:      n,
				}
			}
			n = n.r
		default:
			panic("unreachable: unexpected bit")
		}
	}

	n.masked = true
	n.l, n.r = nil, nil
	n.pruneToRoot()
}

// List returns all prefixes in the tree.
// Overlapping prefixes are merged.
// It will also collapse the neighboring prefixes.
// The order of the prefixes in the output is guaranteed.
//
// Example:
//   - [192.168.0.0/16, 192.168.1.0/24, 192.168.0.1/32] -> [192.168.0.0/16]
//   - [192.168.0.0/32, 192.168.0.1/32] -> [192.168.0.0/31]
func (t *prefixTree) List() []netip.Prefix {
	var (
		rv []netip.Prefix
		q  = []*prefixTreeNode{t.root}
	)

	for len(q) > 0 {
		n := q[len(q)-1]
		q = q[:len(q)-1]

		if n.masked {
			rv = append(rv, n.prefix)
			continue
		}

		if n.l != nil {
			q = append(q, n.l)
		}
		if n.r != nil {
			q = append(q, n.r)
		}
	}

	return rv
}
