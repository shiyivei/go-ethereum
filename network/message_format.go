package network

import (
	"go-publicChain/block"
)

type Version struct {
	Version    int
	BestHeight int64
	AddrFrom   string
}

type GetData struct {
	AddFrom string
	Type    string
	Hash    []byte
}

type Inv struct {
	AddFrom string
	Type    string
	Hashes  [][]byte
}

type Tx struct {
	AddFrom     string
	Transaction []byte
}

type GetBlocks struct {
	AddFrom string
}

type BlockData struct {
	AddFrom string
	Block   *block.Block
}
