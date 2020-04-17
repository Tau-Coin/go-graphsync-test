package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
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
	errRootSize		= errors.New("amt root format error")
	errRootType		= errors.New("amt root type error")
	errAMTNodeType		= errors.New("amt node type error")
	errAMTValNotFound	= errors.New("amt value not found")
)

const (
	AMTSize = 1000
	width = 8

	amtNodeLinksFieldIndex	= 1
	amtNodeValuesFieldIndex	= 2

	valueSize = 1 << 20 //27
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

func amtValueSizeTest() {
	iNode, _:= ipfs.Node()
	root := amt.NewAMT(amt.WrapBlockstore(iNode.Blockstore))

	start := time.Now()

	randBytes := generateRandValue()
	root.Set(0, randBytes)
	cid, err := root.Flush()
	if err != nil {
		fmt.Println("amt flush err:", err)
	}

	fmt.Println("amt large vlue testing took:", time.Since(start))
	fmt.Println("amt root:", cid)
}

func generateRandValue() []byte {
	buf := make([]byte, valueSize)
	rand.Read(buf)
	return buf
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

        amtCtx.Start()
}

func (amtCtx *amtTestContext) Start() {
	start := time.Now()

	height, count, node, err := amtCtx.getAMTRoot(amtCtx.gsCtx.root)
	if err != nil {
		fmt.Println("get amt root err:", err)
		return
	}
	gsFirst := time.Now()

	h, _ := height.AsInt()
	c, _ := count.AsInt()
	fmt.Printf("h:%d, c:%d\n", h, c)
	if amtCtx.index > uint64(c) {
		fmt.Println("index out of range")
		return
	}

	// Then traversal amt dag path for amtCtx.index
	path := []int{}
	_ = travelsal(h, amtCtx.index, &path)
	fmt.Println("amt path:", path)

	// get link
	link, err := getLink(node, path[0])
	if err != nil {
		fmt.Println("get link err:", err)
		return
	}
	fmt.Println("cid link:", link.String())

	selector := amtNodeSelector(path)

	gsSecond0 := time.Now()
	var value string
	err = amtCtx.getValue(link, selector, func(r *result) error {
		if r != nil {
			value = r.val
			fmt.Printf("index:%d, value:%s\n", amtCtx.index, string(value))
			return nil
		}

		return errAMTValNotFound
	})
	if err != nil {
		fmt.Println("amt get value err:", err)
	}
	gsSecond1 := time.Now()

	fmt.Printf("amt graphsync took:%v, first graphsync took:%v, second graphsync:%v\n",
		time.Since(start), gsFirst.Sub(start), gsSecond1.Sub(gsSecond0))
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

func (amtCtx *amtTestContext) getValue(link ipld.Link, selector ipld.Node, cb func(*result) error) error {
	progressChan, errChan := amtCtx.gsCtx.graphExchanger.Request(amtCtx.gsCtx.ctx, amtCtx.pid, link, selector, amtCtx.gsCtx.extension)

	responses := collectResponses(amtCtx.gsCtx.ctx,  progressChan)
        errs := collectErrors(amtCtx.gsCtx.ctx, errChan)
        if len(errs) != 0 {
                for _, e := range errs {
                        fmt.Printf("amt value graphsync error:%v\n", e)
                }
                return errs[0]
        }

	// parse last node
	fmt.Println("response node length:", len(responses))
	valNode := responses[len(responses) - 1].Node
	fmt.Println("last node type:", valNode.ReprKind())
	if valNode.ReprKind() != ipld.ReprKind_String {
		return errAMTValNotFound
	}

	val, err := valNode.AsString()
	if err != nil {
		return err
	}

	cb(&result{val: val})
	return nil
}

type result struct {
	val string
}

func getLink(node ipld.Node, idx int) (ipld.Link, error) {
	if node.ReprKind() != ipld.ReprKind_List {
		return nil, errNodeType
	}

	links, err := node.LookupIndex(amtNodeLinksFieldIndex)
	if err != nil {
		return nil, err
	}

	if links.ReprKind() != ipld.ReprKind_List {
		return nil, errNodeType
	}

	link, err := links.LookupIndex(idx)
	if err != nil {
		return nil, err
	}
	if link.ReprKind() != ipld.ReprKind_Link {
		return nil, errAMTNodeType
	}

	return link.AsLink()
}

func travelsal(height int, i uint64, path *[]int) error {

	if height == 0 {
		*path = append(*path, int(i))
		return nil
	}

	subi := i / nodesForHeight(width, height)
	*path = append(*path, int(subi))

	return travelsal(height - 1, i % nodesForHeight(width, height), path)

}

// AMT root selector should return thress ipld nodes in a list. The first entry
// is the amt height. And the second entry is the amt count. The third entry is the cidlink list.
func amtRootSelector() ipld.Node {
        ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
        return ssb.ExploreAll(ssb.Matcher()).Node()
}

func amtNodeSelector(path []int) ipld.Node {
	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())

	// build value selector
	valIdx := path[len(path) - 1]
	spec := ssb.ExploreIndex(amtNodeValuesFieldIndex, ssb.ExploreIndex(valIdx, ssb.Matcher()))

	// build link selector
	i := len(path) - 2
	for i > 0 {
		idx := path[i]
		spec = ssb.ExploreIndex(amtNodeLinksFieldIndex, ssb.ExploreIndex(idx, spec))

		i--
	}

	selector := spec.Node()
	printSelector(selector)

	return selector
}

func printSelector(selector ipld.Node) {
	str := nodeToJson(selector)
	fmt.Println("selector:", str)
}

func nodesForHeight(width, height int) uint64 {
	val := math.Pow(float64(width), float64(height))
	if val >= float64(math.MaxUint64) {
		fmt.Println("nodesForHeight overflow! This should never happen, please report this if you see this log message")
		return math.MaxUint64
	}

	return uint64(val)
}

