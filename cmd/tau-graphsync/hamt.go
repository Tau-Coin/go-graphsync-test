package main

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"

	_  "github.com/ipfs/go-graphsync"
	ipld "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"github.com/libp2p/go-libp2p-core/peer"

	cbg "github.com/whyrusleeping/cbor-gen"
)

const defaultBitWidth = 8

var (
	errNodeSize	= errors.New("graphsync return nodes' size mismatch")
	errNodeType	= errors.New("graphsync return nodes' type mismatch")
	errNotFound	= errors.New("hamt value not found")
	errHamtFormat	= errors.New("hamt data format error")
)

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

func (hamtCtx *hamtTestContext) Start() {
	hamtCtx.getValue(&hashBits{b: hash(hamtCtx.account)}, hamtCtx.account, hamtCtx.gsCtx.root, func(*KV) error {return nil})
}

func (hamtCtx *hamtTestContext) getValue(hv *hashBits, k string, link ipld.Link, cb func(*KV) error) error {
	idx, err := hv.Next(defaultBitWidth)
	if err != nil {
		return err
	}

	// first of all get bitmap and hash array
	bitmap, array, err := hamtCtx.getHamtNode(link)
	if err != nil {
		return err
	}
	bitfield := parseBitmap(bitmap)
	if bitfield.Bit(idx) == 0 {
		return errNotFound
	}
	cindex := indexForBitPos(idx, bitfield)
	child, err := getChild(cindex, array)
	if err != nil {
		return err
	}
	if child == nil {
		return errNotFound
	}
	fmt.Println("child node type:", child.ReprKind())
	// For hamt data format, child must be map and length must be '1'
	if child.ReprKind() !=ipld.ReprKind_Map && child.Length() != 1 {
		fmt.Printf("%v\n", errHamtFormat)
		return errHamtFormat
	}

	iter := child.MapIterator()
	for !iter.Done() {
		key, value, err := iter.Next()
		if err != nil {
			fmt.Printf("map iter error:%v\n", err)
			return err
		}
		fmt.Printf("key type:%v, value type:%v\n", key.ReprKind(), value.ReprKind())

		if value.ReprKind() == ipld.ReprKind_Link {
			childLink, _ := value.AsLink()
			fmt.Println("cid link:", childLink.String())
			return hamtCtx.getValue(hv, k, childLink, cb)
		}

		// Here the value must be list
		if value.ReprKind() != ipld.ReprKind_List {
			fmt.Println("hamt data format error")
			return errHamtFormat
		}
		listIter := value.ListIterator()
		for !listIter.Done() {
			_, v, err := listIter.Next()
			if err != nil {
				fmt.Printf("list iter error:%v\n", err)
				return err
			}

			// Here the value should be list
			if v.ReprKind() != ipld.ReprKind_List {
				fmt.Println("hamt kv data foramt error")
				return errHamtFormat
			}

			// KV list length should be 2
			if v.Length() != 2 {
				fmt.Printf("kv value node size error:%d\n", v.Length())
				return errHamtFormat
			}

			targetK, _ := value.LookupIndex(0)
			targetV, _ := value.LookupIndex(1)
			fmt.Printf("targetK type:%s, targetV type:%s\n", targetK.ReprKind(), targetV.ReprKind())
			if targetK.ReprKind() == ipld.ReprKind_String && targetV.ReprKind() == ipld.ReprKind_String {
				keyStr, _ := targetK.AsString()
				valueStr, _ := targetV.AsString()
				fmt.Printf("key str:%s, value str:%s\n", keyStr, valueStr)
			}
		}
	}

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

	// responses length should be always 3
	fmt.Printf("reponses nodes size:%d\n", len(responses))
	if len(responses) != 3 {
		fmt.Println("hamt selector error")
		return nil, nil, errNodeSize
	}

	for _, r := range responses {
		fmt.Printf("type:%s, %v\n", r.Node.ReprKind(), r)
		if r.Node.ReprKind() == ipld.ReprKind_List {
			fmt.Println("list length:", r.Node.Length())
		}
		if r.Node.ReprKind() == ipld.ReprKind_Bytes {
			b, err := r.Node.AsBytes()
			fmt.Printf("bytes:%v, err:%v\n", b, err)
		}
	}

	bitmap = responses[1].Node
	array  = responses[2].Node

	if bitmap.ReprKind() != ipld.ReprKind_Bytes || array.ReprKind() != ipld.ReprKind_List {
		fmt.Println("Error:hamt nodes type mismatch")
		return nil, nil, errNodeType
	}

	return bitmap, array, nil
}

func parseBitmap(node ipld.Node) *big.Int {
	bytes, _ := node.AsBytes()
	return big.NewInt(0).SetBytes(bytes)
}

func getChild(idx int, list ipld.Node) (ipld.Node, error) {
	return list.LookupIndex(idx)
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

func indexForBitPos(bp int, bitfield *big.Int) int {
	var x uint
	var count, i int
	w := bitfield.Bits()
	for x = uint(bp); x > bits.UintSize && i < len(w); x -= bits.UintSize {
		count += bits.OnesCount(uint(w[i]))
		i++
	}
	if i == len(w) {
		return count
	}
	return count + bits.OnesCount(uint(w[i])&((1<<x)-1))
}
