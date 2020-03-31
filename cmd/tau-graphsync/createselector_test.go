package main

import (

	"fmt"

	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"github.com/spaolacci/murmur3"

	ipld     "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	selectorbuilder "github.com/ipld/go-ipld-prime/traversal/selector/builder"

)

// state hamt tree
func hamtTreeSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
	//return ssb.ExploreAll(ssb.Matcher()).Node()
	/*
	return ssb.ExploreIndex(1, ssb.ExploreFields(func(efsb selectorbuilder.ExploreFieldsSpecBuilder) {
		//efsb.Insert("0", ssb.ExploreAll(ssb.Matcher()))
		//efsb.Insert("0", ssb.Matcher())
	})).Node()
	*/

	return ssb.ExploreIndex(1, ssb.ExploreAll(
			ssb.ExploreFields(func(efsb selectorbuilder.ExploreFieldsSpecBuilder) {
				efsb.Insert("0", ssb.ExploreAll(ssb.Matcher()))
				//efsb.Insert("0", ssb.Matcher())
		}))).Node()

}

// state hamt tree
func accountStateSelector(cidLink bool, cindex int) ipld.Node {

	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())

	if cidLink {
		return ssb.ExploreIndex(0, ssb.Matcher()).Node()
	} else {
		return ssb.ExploreIndex(1, ssb.ExploreIndex(cindex, ssb.Matcher())).Node()
	}
}

// map 
func mapSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
    return ssb.ExploreAll(ssb.Matcher()).Node()
}
