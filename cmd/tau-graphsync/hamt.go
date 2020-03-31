package main

import (
	_  "github.com/ipfs/go-graphsync"
	_  "github.com/ipld/go-ipld-prime"
	"github.com/libp2p/go-libp2p-core/peer"
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

func (hamtCtx *hamtTestContext) Start() {}
