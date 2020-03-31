package main

import (
	"fmt"

	_  "github.com/ipfs/go-graphsync"
	ipld "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"github.com/libp2p/go-libp2p-core/peer"

	cbg "github.com/whyrusleeping/cbor-gen"
)

const defaultBitWidth = 8

type hamtTestContext struct {
	gsCtx			*GraphsyncContext
	pid			peer.ID
	account			string
}

func triggerHamtTest(gsCtx *GraphsyncContext, pid peer.ID, account string) {
	hamtCtx := &hamtTestContext{
		gsCtx:		gsCtx,
		pid:		pid,
		account:	account,
	}

	go hamtCtx.Start()
}

func (hamtCtx *hamtTestContext) Start() {}

func (hamtCtx *hamtTestContext) getValue(hv *hashBits, k string, link ipld.Link, cb func(*KV) error) error {
	_, err := hv.Next(defaultBitWidth)
	if err != nil {
		return err
	}

	// first of all get bitmap and hash array

	return nil
}


//Note return values: the first is the bitmap, and the second is corresponding hash array.
func (hamtCtx *hamtTestContext) getHamtNode(link ipld.Link) (ipld.Node, ipld.Node, error) {
	progressChan, errChan := hamtCtx.gsCtx.graphExchanger.Request(hamtCtx.gsCtx.ctx, hamtCtx.pid, link, hamtSelector(), hamtCtx.gsCtx.extension)

	responses := collectResponses(hamtCtx.gsCtx.ctx,  progressChan)
	errs := collectErrors(hamtCtx.gsCtx.ctx, errChan)
	if len(errs) != 0 {
		for _, e := range errs {
			fmt.Printf("graphsync error:%v\n", e)
		}
		return nil, nil, errs[0]
	}

	var (
		bitmap ipld.Node
		array  ipld.Node
	)

	// responses length should be always 2
	for _, n := range responses {
		fmt.Printf("%v\n", n)
	}

	return bitmap, array, nil
}

type KV struct {
	Key   string
	Value *cbg.Deferred
}

// HAMT selector should return two ipld nodes in a list. The first entry
// is the bitmap. And the second entry is the map.
func hamtSelector() ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
	return ssb.ExploreAll(ssb.Matcher()).Node()
}
