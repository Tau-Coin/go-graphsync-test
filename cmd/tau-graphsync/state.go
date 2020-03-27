package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	ipfs "github.com/ipfs/go-ipfs/lib"
	hamt "github.com/ipfs/go-hamt-ipld"
	ipld "github.com/ipld/go-ipld-prime"
)

func randString() string {
	buf := make([]byte, 20)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func randValue() []byte {
	buf := make([]byte, 16)
	rand.Read(buf)
	return buf
}

func testString(i int) string {
	keyTmp :=  "hamtk"+ string(i)
	return keyTmp
}

func testValue(i int) []byte {
	valueTmp :=  "hamtv"+ string(i)
	return []byte(valueTmp)
}

// setup state hamt tree, and return the root link or some error
func setupStateTree(ctx context.Context) (ipld.Link, error) {

	fmt.Println("Start key and value random assignment")
	vals := make(map[string][]byte)
	var keys []string
	for i := 0; i < 10; i++ {
		//s := randString()
		//vals[s] = randValue()
		s := testString(i)
		vals[s] = testValue(i)
		keys = append(keys, s)
	}

	iNode, _:= ipfs.Node()
	bStore := iNode.Blockstore
	cs := hamt.CSTFromBstore(bStore)
	begn := hamt.NewNode(cs)

	fmt.Println("Start set")
	bSet := time.Now()
	for _, k := range keys {
		if err := begn.Set(ctx, k, vals[k]); err != nil {
			return nil,  err
		}
	}
	fmt.Println("set took: ", time.Since(bSet))

	fmt.Println("start flush")
	bef := time.Now()
	if err := begn.Flush(ctx); err != nil {
		return nil, err
	}
	fmt.Println("flush took: ", time.Since(bef))

	fmt.Println("start Put")
	befPut := time.Now()
	c, err := cs.Put(ctx, begn)
	if err != nil {
		return nil, err
	}
	fmt.Println("Put took: ", time.Since(befPut))

	fmt.Println("Cid is: ", c)

	return cidlink.Link{c}, nil
}
