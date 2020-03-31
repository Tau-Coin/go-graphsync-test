package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p-core/host"

	ipld "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/encoding/dagjson"

	graphsync "github.com/ipfs/go-graphsync"
	graphsyncimpl "github.com/ipfs/go-graphsync/impl"
	"github.com/ipfs/go-graphsync/ipldbridge"
	gsnet "github.com/ipfs/go-graphsync/network"
	storeutil "github.com/ipfs/go-graphsync/storeutil"

	blockstore "github.com/ipfs/go-ipfs-blockstore"
	ipfs "github.com/ipfs/go-ipfs/lib"
	"github.com/libp2p/go-libp2p-core/peer"
)

type GraphsyncContext struct {
	ctx			context.Context

	host			host.Host

	gsnet			gsnet.GraphSyncNetwork

	bridge			ipldbridge.IPLDBridge

	bs			blockstore.Blockstore

	loader			ipld.Loader

	storer			ipld.Storer

	root			ipld.Link

	extensionData		[]byte

	extensionName		graphsync.ExtensionName

	extension		graphsync.ExtensionData

	extensionResponseData	[]byte

	extensionResponse	graphsync.ExtensionData

	graphExchanger		graphsync.GraphExchange
}

func setupGSContext(ctx context.Context, root ipld.Link) (*GraphsyncContext, error) {
	gsCtx := &GraphsyncContext{
		ctx:	ctx,
		root:	root,
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

	gsCtx.graphExchanger		= gsCtx.GraphSyncHost()

	return gsCtx, nil
}

func (gsCtx *GraphsyncContext) GraphSyncHost() graphsync.GraphExchange {
	return graphsyncimpl.New(gsCtx.ctx, gsCtx.gsnet, gsCtx.bridge, gsCtx.loader, gsCtx.storer)
}

func (gsCtx *GraphsyncContext) Host() host.Host {
	return gsCtx.host
}

func (gsCtx *GraphsyncContext) GraphsyncTest(pid peer.ID, account string) {
	var (
		receivedResponseData []byte
		receivedRequestData []byte
	)

	err := gsCtx.graphExchanger.RegisterResponseReceivedHook(
		func(p peer.ID, responseData graphsync.ResponseData) error {
			fmt.Println("graphsync response received...")
			data, has := responseData.Extension(gsCtx.extensionName)
			if has {
				receivedResponseData = data
				fmt.Println("reponse extension ", receivedResponseData)
			}
			fmt.Println("reponse status:", responseData.Status())
			return nil
		})
	if err != nil {
		fmt.Printf("Error setting up extension:%v\n", err)
		return
	}

	err = gsCtx.graphExchanger.RegisterRequestReceivedHook(func(p peer.ID, requestData graphsync.RequestData, hookActions graphsync.RequestReceivedHookActions) {
		fmt.Printf("graphsync reqeust received, root:%v, selector:%v\n", requestData.Root(), requestData.Selector())
		var has bool
		receivedRequestData, has = requestData.Extension(gsCtx.extensionName)
		if !has {
			hookActions.TerminateWithError(errors.New("Missing extension"))
		} else {
			hookActions.SendExtensionData(gsCtx.extensionResponse)
		}
	})

	if err != nil {
		fmt.Printf("Error setting up extension:%v\n", err)
		return
	}

	fmt.Printf("graphsync is starting..., sync accout is: %v\n", account)
	//start := time.Now()

	//progressChan, errChan := gsCtx.graphExchanger.Request(gsCtx.ctx, pid, gsCtx.root, mapSelector(), gsCtx.extension)
	//progressChan, errChan := gsCtx.graphExchanger.Request(gsCtx.ctx, pid, gsCtx.root, hamtTreeSelector(), gsCtx.extension)
	//progressChan, errChan := gsCtx.graphExchanger.Request(gsCtx.ctx, pid, gsCtx.root, accountStateSelector(account), gsCtx.extension)

	//responses := collectResponses(gsCtx.ctx,  progressChan)
	//errs := collectErrors(gsCtx.ctx, errChan)

	//fmt.Printf("graphsyn result, response size:%d, errors size:%d\n", len(responses), len(errs))
	//if len(errs) != 0 {
		//fmt.Println("errors during traverse")
		//return
	//}

	//fmt.Println("graph sync took: ", time.Since(start))

	/*
	for _, response := range responses {
		fmt.Printf("node path:%s\n", response.Path.String())
		//fmt.Printf("node json:%s\n", nodeToJson(response.Node))
	}
	*/

	triggerHamtTest(gsCtx, pid, account)
}

// collectResponses is just a utility to convert a graphsync response progress
// channel into an array.
func collectResponses(ctx context.Context, responseChan <-chan graphsync.ResponseProgress) []graphsync.ResponseProgress {
	var collectedBlocks []graphsync.ResponseProgress
	for {
		select {
		case blk, ok := <-responseChan:
			if !ok {
				return collectedBlocks
			}
			collectedBlocks = append(collectedBlocks, blk)
		case <-ctx.Done():
			fmt.Println("response channel never closed")
		}
	}
}

// collectErrors is just a utility to convert an error channel into an array.
func collectErrors(ctx context.Context, errChan <-chan error) []error {
	var collectedErrors []error
	for {
		select {
		case err, ok := <-errChan:
			if !ok {
				return collectedErrors
			}
			fmt.Printf("graphsync err:%v\n", err)
			collectedErrors = append(collectedErrors, err)
		case <-ctx.Done():
			fmt.Println("error channel never closed")
		}
	}
}

func nodeToJson(node ipld.Node) string {
	buff := new(bytes.Buffer)
	err := dagjson.Encoder(node, buff)
	if err != nil {
		fmt.Println("dagjson encode err:", err)
		return ""
	} else {
		return buff.String()
	}
}
