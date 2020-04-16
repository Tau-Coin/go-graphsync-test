package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	amt "github.com/filecoin-project/go-amt-ipld"
	//cid "github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	ipfs "github.com/ipfs/go-ipfs/lib"
	ipld "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"github.com/libp2p/go-libp2p-core/peer"
)

var (
	errRootSize = errors.New("amt root format error")
	errRootType = errors.New("amt root type error")
)

const (
	AMTSize = 1000
	width = 8
)

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

type amtTestContext struct {
	gsCtx	*GraphsyncContext
	pid	peer.ID
	index	uint64
}

func triggerAMTTest(gsCtx *GraphsyncContext, pid peer.ID, index string) {
	idxInt64, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		fmt.Println("parse int err:", err)
		return
	}

	amtCtx := &amtTestContext{
		gsCtx:	gsCtx,
		pid:	pid,
		index:	uint64(idxInt64),
        }

        go amtCtx.Start()
}

func (amtCtx *amtTestContext) Start() {
	start := time.Now()

	height, count, _, err := amtCtx.getAMTRoot(amtCtx.gsCtx.root)
	if err != nil {
		fmt.Println("get amt root err:", err)
		return
	}

	h, _ := height.AsInt()
	c, _ := count.AsInt()
	fmt.Printf("h:%d, c:%d\n", h, c)

	fmt.Println("amt graphsync took:", time.Since(start))
}

func (amtCtx *amtTestContext) getAMTRoot(link ipld.Link) (ipld.Node, ipld.Node, ipld.Node, error) {
        progressChan, errChan := amtCtx.gsCtx.graphExchanger.Request(amtCtx.gsCtx.ctx, amtCtx.pid, link, amtRootSelector(), amtCtx.gsCtx.extension)

        responses := collectResponses(amtCtx.gsCtx.ctx,  progressChan)
        errs := collectErrors(amtCtx.gsCtx.ctx, errChan)
        if len(errs) != 0 {
                for _, e := range errs {
                        fmt.Printf("graphsync error:%v\n", e)
                }
                return nil, nil, nil, errs[0]
        }

        var (
		height	ipld.Node
		count	ipld.Node
                node	ipld.Node
        )

	// responses length should be always 4
        fmt.Printf("reponses nodes size:%d\n", len(responses))
        if len(responses) != 4 {
                fmt.Println("amt selector error")
                return nil, nil, nil, errRootSize
        }

	height	= responses[1].Node
        count	= responses[2].Node
	node	= responses[3].Node

        if height.ReprKind() != ipld.ReprKind_Int || count.ReprKind() != ipld.ReprKind_Int || node.ReprKind() != ipld.ReprKind_List {
                fmt.Println("Error:amt nodes type mismatch")
                return nil, nil, nil, errRootType
        }

        return height, count, node, nil
}

// AMT root selector should return thress ipld nodes in a list. The first entry
// is the amt height. And the second entry is the amt count. The third entry is the cidlink list.
func amtRootSelector() ipld.Node {
        ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
        return ssb.ExploreAll(ssb.Matcher()).Node()
}
