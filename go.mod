module github.com/Tau-Coin/go-graphsync-test

go 1.14

require (
	github.com/filecoin-project/go-amt-ipld v1.0.0
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-cid v0.0.5
	github.com/ipfs/go-graphsync v0.0.4
	github.com/ipfs/go-hamt-ipld v0.0.0-00010101000000-000000000000
	github.com/ipfs/go-ipfs v0.4.23
	github.com/ipfs/go-ipfs-blockstore v0.1.4
	github.com/ipfs/go-ipld-cbor v0.0.4
	github.com/ipfs/interface-go-ipfs-core v0.2.6
	github.com/ipld/go-ipld-prime v0.0.2-0.20191108012745-28a82f04c785
	github.com/libp2p/go-eventbus v0.1.0
	github.com/libp2p/go-libp2p-core v0.5.0
	github.com/multiformats/go-multiaddr v0.2.1
	github.com/multiformats/go-multihash v0.0.13
	github.com/spaolacci/murmur3 v1.1.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20200123233031-1cdf64d27158
)

replace github.com/ipfs/go-ipfs => github.com/tau-coin/go-ipfs v0.4.22-0.20200313092758-7b227442e904

replace github.com/ipfs/go-hamt-ipld => github.com/ipfs/go-hamt-ipld v0.0.13
