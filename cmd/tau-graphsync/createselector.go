package main

import (
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	ipld     "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	ipldselector "github.com/ipld/go-ipld-prime/traversal/selector"
	//selectorbuilder "github.com/ipld/go-ipld-prime/traversal/selector/builder"
)

// state hamt tree
func stateSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
    //return ssb.ExploreAll(ssb.Matcher()).Node()
    /*
    return ssb.ExploreIndex(1, ssb.ExploreFields(func(efsb selectorbuilder.ExploreFieldsSpecBuilder) {
		//efsb.Insert("0", ssb.ExploreAll(ssb.Matcher()))
		//efsb.Insert("0", ssb.Matcher())
		efsb.Insert("0", ssb.ExploreFields(func(efsb1 selectorbuilder.ExploreFieldsSpecBuilder) {
			efsb1.Insert("/", ssb.ExploreAll(ssb.Matcher()))
		}))
	})).Node()
	*/

	return ssb.ExploreRecursive(ipldselector.RecursionLimitNone(),
		ssb.ExploreAll(ssb.ExploreRecursiveEdge())).Node()
}

// map 
func mapSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
    return ssb.ExploreAll(ssb.Matcher()).Node()
}
