package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ipld/go-ipld-prime/fluent"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	hamt "github.com/ipfs/go-hamt-ipld"
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
			//mb.Insert(knb.CreateString("Coin"), vnb.CreateString("TAU"))
			mb.Insert(knb.CreateString("hamt"), vnb.CreateString("bafy2bzaceciimlk7tioay4g2zd226l35s4g3gu6fuwfd6l53nddv3nra4b6xw"))
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

func readKeyValue()([]string, map[string][]byte){
	file, err := os.Open("./testdata/taudata.txt")
	if err != nil {
		fmt.Println("Open tau data failed, err: ", err)
	}
	defer file.Close()

	var keys []string
	vals := make(map[string][]byte)

	reader := bufio.NewReader(file)
	for {
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		//split
		strContent := strings.Split(str, " ")

		strQuitLF := strings.Replace(strContent[1], "\n", "", -1)
		//keys, vals
		//fmt.Println(strContent[0], strQuitLF)
		keyTmp := strContent[0]
		valueTmp := []byte(strQuitLF)
		vals[keyTmp]= valueTmp
		keys= append(keys, keyTmp)
	}
	return keys, vals
}

// setup state hamt tree, and return the root link or some error
func createStateNode(ctx context.Context) (ipld.Link, error) {

	fmt.Println("Start reading file")
	keys, vals := readKeyValue()
	fmt.Println("End reading file")

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

	for i := 0; i < 100000000; i++ {
		if err := begn.Set(ctx, randString(), randValue()); err != nil {
			fmt.Println("hamt node set err:", err)
			return nil, err
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

	hamtTest(ctx, cs, c)

	return cidlink.Link{c}, nil
}

var testKeys = []string {
	"TA7QFFvxF1nLZUG4ecp685dLbkJoBZwUAm",
	"TA7XKuSZBGwFds8YHhN2Wmkm2BK7qDpGKb",
	"TA83nWtuYPmjdGexR3nZ2RSzXorGyw4BTz",
	"TA8GswGDaMVMZVVSB3CETR9Xj7y1urDr59",
	"TA8hdCdURvDfzL2fCJUCS9TM1NTmsGZuid",
	"TA8PUrT842mEgCyKg9S3FDDDktcR99Lch7",
	"TA8zQhMvQuzvBjfc1EzjFAH7TLX2fJ9K7S",
	"TA98HVHB1r1GLu5vxp7TDC6szfBiS2dK5K",
	"TA98kjzQekCkwXsmeUvxdfGaVJdHWzCaVF",
	"TA9b2PmZukevtPzKWYdZSFaRmmJXpgdZEw",
}

func hamtTest(ctx context.Context, cs *hamt.CborIpldStore, c cid.Cid) {
	node := hamt.NewNode(cs)
	err := cs.Get(ctx, c, node)
	if err != nil {
		fmt.Println("block store get hamt node err:", err)
		return
	}

	costs	:= make(map[string]time.Duration)
	results	:= make(map[string]string)

	for _, k := range testKeys {
		start := time.Now()
		var v []byte
		err := node.Find(ctx, k, &v)
		if err == nil {
			costs[k] = time.Since(start)
			results[k] = string(v)
		}
	}

	for key, value := range results {
		fmt.Printf("key:%s, value:%s, time cost:%v\n", key, value, costs[key])
	}
}
