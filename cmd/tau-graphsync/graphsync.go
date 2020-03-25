package main

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"

	ipld "github.com/ipld/go-ipld-prime"

	graphsync "github.com/ipfs/go-graphsync"
	graphsyncimpl "github.com/ipfs/go-graphsync/impl"
	"github.com/ipfs/go-graphsync/ipldbridge"
	gsnet "github.com/ipfs/go-graphsync/network"
	storeutil "github.com/ipfs/go-graphsync/storeutil"

	blockstore "github.com/ipfs/go-ipfs-blockstore"
	ipfs "github.com/ipfs/go-ipfs/lib"
)

type GraphsyncContext struct {
	ctx			context.Context

	host			host.Host

	gsnet			gsnet.GraphSyncNetwork

	bridge			ipldbridge.IPLDBridge

	bs			blockstore.Blockstore

	loader			ipld.Loader

	storer			ipld.Storer

	extensionData		[]byte

	extensionName		graphsync.ExtensionName

	extension		graphsync.ExtensionData

	extensionResponseData	[]byte

	extensionResponse	graphsync.ExtensionData
}

func setupGSContext(ctx context.Context) (*GraphsyncContext, error) {
	gsCtx := &GraphsyncContext{
		ctx: ctx,
	}

	// get ipfs node
	ipfsNode, err := ipfs.Node()
	if err != nil {
		return gsCtx, err
	}

	gsCtx.host		= ipfsNode.PeerHost
	gsCtx.gsnet		= gsnet.NewFromLibp2pHost(gsCtx.host)
	gsCtx.bridge		= ipldbridge.NewIPLDBridge()
	gsCtx.bs		= ipfsNode.Blockstore
	gsCtx.loader		= storeutil.LoaderForBlockstore(gsCtx.bs)
	gsCtx.storer		= storeutil.StorerForBlockstore(gsCtx.bs)
	gsCtx.extensionData	= []byte("taucoin-graphsync-test-request")
	gsCtx.extensionName	= graphsync.ExtensionName("taucoin")
	gsCtx.extension		= graphsync.ExtensionData{
		Name: gsCtx.extensionName,
		Data: gsCtx.extensionData,
	}
	gsCtx.extensionResponseData	= []byte("taucoin-graphsync-test-response")
	gsCtx.extensionResponse		= graphsync.ExtensionData{
		Name: gsCtx.extensionName,
		Data: gsCtx.extensionResponseData,
	}

	return gsCtx, nil
}

func (gsCtx *GraphsyncContext) GraphSyncHost() graphsync.GraphExchange {
	return graphsyncimpl.New(gsCtx.ctx, gsCtx.gsnet, gsCtx.bridge, gsCtx.loader, gsCtx.storer)
}

func (gsCtx *GraphsyncContext) Host() host.Host {
	return gsCtx.host
}
