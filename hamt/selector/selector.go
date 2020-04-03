package selector

// This file is all about hamt selector.
// Please refer to https://github.com/ipfs/go-hamt-ipld

import (
	"errors"
	"fmt"
	"math/big"

	ipld "github.com/ipld/go-ipld-prime"
	ipldselector "github.com/ipld/go-ipld-prime/traversal/selector"

	"github.com/Tau-Coin/go-graphsync-test/hamt"
)

const DefaultBitWidth = 8

var (
        ErrNodeSize     = errors.New("graphsync return nodes' size mismatch")
        ErrNodeType     = errors.New("graphsync return nodes' type mismatch")
        ErrNotFound     = errors.New("hamt value not found")
        ErrHamtFormat   = errors.New("hamt data format error")
)

type HamtContext struct {

	// The key and its hash bits
	key		string
	hashBits	*hamt.HashBits
	bitWidth	int

	// Hamt node bitmap
	bitfield	*big.Int
	index		int
}

type HamtSelector struct {
	ctx	*HamtContext
	value	[]byte
	err	error
}

func New(key string, bitWidth int) *HamtSelector {
	return &HamtSelector{
		ctx:	&HamtContext{
				key:		key,
				hashBits:	&hamt.HashBits{B: hamt.Hash(key)},
				bitWidth:	bitWidth,
				index:		-1,
		},
	}
}

// Retrun the hamt node pointer array index.
func (hs *HamtSelector) Interests() []ipld.PathSegment {
	return []ipld.PathSegment{ipld.PathSegmentOfInt(1)}
}

func (hs *HamtSelector) Explore(node ipld.Node, ps ipld.PathSegment) ipldselector.Selector {
	// First of all, retrive the bitmap
	if err := hs.retriveBitmap(node); err != nil {
		hs.err = err
		return nil
	}

	if err := hs.computePointerIndex(); err != nil {
		hs.err = err
		return nil
	}

	fmt.Println("pointer index:", hs.ctx.index)

	// At last, return PointerSelector
	return NewPointerSelector(hs)
}

func (hs *HamtSelector) Decide(n ipld.Node) bool {
	return false
}

func (hs *HamtSelector) retriveBitmap(n ipld.Node) error {
        if n.ReprKind() != ipld.ReprKind_Link && n.Length() != 2 {
                fmt.Printf("Hamt selector err:%v\n", ErrNodeSize)
                return ErrNodeSize
        }

        bitmapNode, err := n.LookupIndex(0)
        if err != nil {
                fmt.Printf("lookup bitmap err:%v\n", err)
                return err
        }
        if bitmapNode == nil || bitmapNode.ReprKind() != ipld.ReprKind_Bytes {
                fmt.Println("bitmap bytes not found")
                return ErrHamtFormat
        }

	bitmapBytes, err := bitmapNode.AsBytes()
	if err != nil {
		fmt.Printf("retrive bitmap bytes err:%v\n", err)
		return err
	}

	hs.ctx.bitfield = big.NewInt(0).SetBytes(bitmapBytes)

	return nil
}

func (hs *HamtSelector) computePointerIndex() error {
	idx, err := hs.ctx.hashBits.Next(hs.ctx.bitWidth)
	if err != nil {
		fmt.Printf("compute pointer index err:%v\n", err)
		return err
	}
	if hs.ctx.bitfield.Bit(idx) == 0 {
		return ErrNotFound
	}

	index := hamt.IndexForBitPos(idx, hs.ctx.bitfield)
	if index < 0 || index > 255 {
		hs.err = errors.New(fmt.Sprintf("invalid pointer index:%d", index))
		return hs.err
	}
	hs.ctx.index = index

	return nil
}

func (hs *HamtSelector) GetResult() ([]byte, error) {
	return hs.value, hs.err
}

// Hamt pointer select which in fact is a index selector.
type PointerSelector struct {
	hs	*HamtSelector
}

func NewPointerSelector(hamtSelector *HamtSelector) *PointerSelector {
	return &PointerSelector{hs: hamtSelector}
}

func (ps *PointerSelector) Interests() []ipld.PathSegment {
	return []ipld.PathSegment{ipld.PathSegmentOfInt(ps.hs.ctx.index)}
}

func (ps *PointerSelector) Explore(node ipld.Node, pseg ipld.PathSegment) ipldselector.Selector {
	// Ensure node is a ipld list
	if node.ReprKind() != ipld.ReprKind_List {
		fmt.Println("pointer selector err: node isn't ipld list")
		ps.hs.err = ErrHamtFormat
		return nil
	}

	return NewEntrySelector(ps.hs)
}

func (ps *PointerSelector) Decide(n ipld.Node) bool {
	return false
}

type EntrySelector struct {
	hs	*HamtSelector
}

func NewEntrySelector(hamtSelector *HamtSelector) *EntrySelector {
	return &EntrySelector{hs: hamtSelector}
}

func (es *EntrySelector) Interests() []ipld.PathSegment {
	return []ipld.PathSegment{ipld.PathSegmentOfString("0")}
}

func (es *EntrySelector) Explore(node ipld.Node, ps ipld.PathSegment) ipldselector.Selector {
	// Ensure node is a ipld map
	if node.ReprKind() != ipld.ReprKind_List {
		fmt.Println("entry selector err: node isn't ipld map")
		es.hs.err = ErrHamtFormat
		return nil
	}

	// Judge whether value is a link or a key-value array.
	linkNode, _ := node.LookupString("0")
	kvArrayNode, _ := node.LookupString("1")
	if linkNode == nil || kvArrayNode == nil {
		fmt.Println("entry selector err:", ErrHamtFormat)
		es.hs.err = ErrHamtFormat
		return nil
	}
	if linkNode != nil && kvArrayNode != nil {
		fmt.Println("entry selector err:", ErrHamtFormat)
		es.hs.err = ErrHamtFormat
		return nil
	}

	if linkNode != nil && linkNode.ReprKind() == ipld.ReprKind_Link {
		link, err := linkNode.AsLink()
		if err == nil {
			fmt.Println("cid link: ", link.String())
		}

		// recursive hamt selector
		return es.hs
	}

	if kvArrayNode != nil && kvArrayNode.ReprKind() == ipld.ReprKind_List {
		// compare the key value
		listIter := kvArrayNode.ListIterator()
		for !listIter.Done() {
			_, v, err := listIter.Next()
			if err != nil {
				fmt.Printf("list iter error:%v\n", err)
				es.hs.err = err
				return nil
			}

			// Here the value should be list
			if v.ReprKind() != ipld.ReprKind_List {
				fmt.Println("hamt kv data foramt error")
				es.hs.err = ErrHamtFormat
				return nil
                        }

                        // KV list length should be 2
                        if v.Length() != 2 {
                                fmt.Printf("kv value node size error:%d\n", v.Length())
                                es.hs.err = ErrHamtFormat
				return nil
                        }

                        targetK, _ := v.LookupIndex(0)
                        targetV, _ := v.LookupIndex(1)
                        //fmt.Printf("targetK type:%s, targetV type:%s\n", targetK.ReprKind(), targetV.ReprKind())
                        if targetK.ReprKind() == ipld.ReprKind_String && targetV.ReprKind() == ipld.ReprKind_Bytes {
				keyStr, _ := targetK.AsString()
				valueBytes, _ := targetV.AsBytes()
				if keyStr == es.hs.ctx.key {
					fmt.Println("Got the result for the key:", keyStr)
					es.hs.value = valueBytes
					es.hs.err = nil
					return nil
				}
			}
		}

		es.hs.err = ErrNotFound
		return nil
	}

	es.hs.err = ErrNotFound
	return nil
}

func (es *EntrySelector) Decide(n ipld.Node) bool {
	return false
}
