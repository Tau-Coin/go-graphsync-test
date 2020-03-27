package main

import (
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	ipld     "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
)

// state hamt tree
func stateSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
    return ssb.ExploreRecursiveEdge().Node()
}

// map 
func mapSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
    return ssb.ExploreAll(ssb.Matcher()).Node()
}
