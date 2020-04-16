package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	amt "github.com/filecoin-project/go-amt-ipld"
	//cid "github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	ipfs "github.com/ipfs/go-ipfs/lib"
	ipld "github.com/ipld/go-ipld-prime"
	//ipldfree "github.com/ipld/go-ipld-prime/impl/free"
)

const AMTSize = 1000

func createAMTRoot(ctx context.Context) (ipld.Link, error) {
	iNode, _:= ipfs.Node()
	root := amt.NewAMT(amt.WrapBlockstore(iNode.Blockstore))

	start := time.Now()

	i := 0
	for i < AMTSize {
		root.Set(uint64(i), strconv.Itoa(i))
		i++
	}

	cid, err := root.Flush()

	fmt.Println("create amt took:", time.Since(start))
	fmt.Println("amt root:", cid)

	return cidlink.Link{cid}, err
}
