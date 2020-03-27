package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ipld/go-ipld-prime/fluent"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	ipfs "github.com/ipfs/go-ipfs/lib"
	ipld "github.com/ipld/go-ipld-prime"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	mh "github.com/multiformats/go-multihash"
)

// create a map, and return the root link or some error
func createMapNode(ctx context.Context) (ipld.Link, error) {

	fmt.Println("Start to set map content")
	//build node
	mapNodeBuilder := fluent.WrapNodeBuilder(ipldfree.NodeBuilder())
    mapNode := mapNodeBuilder.CreateMap(func(mb fluent.MapBuilder, knb fluent.NodeBuilder, vnb fluent.NodeBuilder) {
			mb.Insert(knb.CreateString("Coin"), vnb.CreateString("TAU"))
		})

	//node -> link
	linkBuilder := cidlink.LinkBuilder{Prefix: cid.NewPrefixV1(cid.DagCBOR, mh.SHA2_256)}
	iNode, _:= ipfs.Node()
	bStore := iNode.Blockstore
	storer := func(lnkCtx ipld.LinkContext) (io.Writer, ipld.StoreCommitter, error) {
        var buffer bytes.Buffer
        committer := func(lnk ipld.Link) error {
			asCidLink, ok := lnk.(cidlink.Link)
			if !ok {
                return fmt.Errorf("Unsupported Link Type")
            }
            block, err := blocks.NewBlockWithCid(buffer.Bytes(), asCidLink.Cid)
            if err != nil {
                return err
            }
            return bStore.Put(block)
        }
        return &buffer, committer, nil
    }

	linkMapNode, err := linkBuilder.Build(ctx, ipld.LinkContext{}, mapNode, storer)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Mapnode's link is: %v\n", linkMapNode.String())

	return linkMapNode, nil
}
