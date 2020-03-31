package main

import (

	"fmt"

	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"github.com/spaolacci/murmur3"

	ipld     "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	selectorbuilder "github.com/ipld/go-ipld-prime/traversal/selector/builder"

)

func hamtHash(val string) []byte{
    h := murmur3.New64()
    h.Write([]byte(val))
    return h.Sum(nil)
}

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
func accountStateSelector(account string) ipld.Node {
	// -> murmur3 hash
	hasher := hamtHash(account)
	fmt.Println("Account's murmur3 hash is: ", hasher)

	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())

	return ssb.ExploreIndex(1, ssb.ExploreAll(
			ssb.ExploreFields(func(efsb selectorbuilder.ExploreFieldsSpecBuilder) {
				efsb.Insert("0", ssb.ExploreAll(ssb.Matcher()))
				//efsb.Insert("0", ssb.Matcher())
		}))).Node()

}

// map 
func mapSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
    return ssb.ExploreAll(ssb.Matcher()).Node()
}
