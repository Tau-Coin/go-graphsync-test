package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/libp2p/go-eventbus"
	"github.com/libp2p/go-libp2p-core/event"
	ipld "github.com/ipld/go-ipld-prime"
	ipfs "github.com/ipfs/go-ipfs/lib"
	util "github.com/ipfs/go-ipfs/lib/util"

	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

var (
	repoDir = "tau-ipfs"
)

func main() {

	/*
	if len(os.Args)!=2{
		fmt.Println("Please check your command!")
		os.Exit(1)
	}
	 */

	account := "1"

	intrh, ctx := util.SetupInterruptHandler(context.Background())
	defer intrh.Close()

	var (
		err		error
		errCh		<-chan error
		wg		sync.WaitGroup
		repoPath	string
		subscription	event.Subscription
		gsCtx		*GraphsyncContext
		root		ipld.Link
	)

	repoPath = os.Getenv("HOME") + "/" + repoDir
	fmt.Printf("repo path is:%s\n", repoPath)

	if err = ipfs.InitIpfs(repoPath); err != nil {
		fmt.Printf("init ipfs error:%v\n", err)
		os.Exit(1)
	}

	// Start ipfs daemon
	err, errCh = ipfs.StartDaemon();
	if err != nil {
		fmt.Printf("start daemon error:%v\n", err)
		os.Exit(1)
	}

	// Here ipfs daemon is running, so run some test cases.
	testCoreAPI()

	// Anyway, we should setup state tree firstly.
	//root, err = createMapNode(ctx)
	//root, err = createStateNode(ctx)
	root, err = createAMTRoot(ctx)
	if err != nil {
		fmt.Printf("setup state tree error:%v\n", err)
		os.Exit(1)
	}
	//amtValueSizeTest()

	// Then setup graphsync context
	gsCtx, err = setupGSContext(ctx, root)
	if err != nil {
		fmt.Printf("setup graphsync error:%v\n", err)
		os.Exit(1)
	}

	subscription, err = gsCtx.Host().EventBus().Subscribe(&event.EvtPeerIdentificationCompleted{}, eventbus.BufSize(32))
	wg.Add(1)
	go handleEvent(wg, subscription, gsCtx, account)

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case derr := <-errCh:
			if derr != ipfs.ErrNormalExit {
				fmt.Printf("ipfs daemon internal error:%v\n", derr)
				ipfs.StopDaemon()
			} else {
				fmt.Println("ipfs daemon exit normally")
			}
		case <-ctx.Done():
			ipfs.StopDaemon()
		}
	}()

	wg.Wait()
}

func testCoreAPI() {
	var api coreiface.CoreAPI
	var err error

	fmt.Println("Coreapi test is starting...")

	if api, err = ipfs.API(); err != nil {
		fmt.Printf("get api error:%v\n", err)
		fmt.Println("Coreapi test failed.")
		return
	}

        keyAPI := api.Key()
        id, _ := keyAPI.Self(context.Background())
        fmt.Printf("IPFS Node id:%s\n", id.ID())

	fmt.Println("Coreapi test passed")
}

func handleEvent(wg sync.WaitGroup, sub event.Subscription, gsCtx *GraphsyncContext, account string) {
	defer wg.Done()

	defer func() {
		_ = sub.Close()
		// drain the channel.
		for range sub.Out() {
		}
	}()

	fmt.Println("starting handle event for testing account: ", account)
	for {
		select {
		case evt, more := <-sub.Out():
			fmt.Printf("receive event:%v\n", evt)
			fmt.Println("event type:", reflect.TypeOf(evt))
			if !more {
				return
			}

			idCompletedEvt, ok := evt.(event.EvtPeerIdentificationCompleted)
			if ok {
				pid := idCompletedEvt.Peer
				fmt.Printf("Identity completed peer:%s\n", pid)
				//if isRelay(gsCtx.Host().Peerstore().Addrs(pid)) {
					// trigger graphsync process
					go gsCtx.GraphsyncTest(pid, account)
				//}
			}
		case <-gsCtx.ctx.Done():
			return
		}
	}
}
