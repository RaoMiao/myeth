package types

import "myeth/common"

//区块头
type Header struct {
	ParentHash common.Hash    `json:"parentHash"       gencodec:"required"`
	UncleHash  common.Hash    `json:"sha3Uncles"       gencodec:"required"`
	Coinbase   common.Address `json:"miner"            gencodec:"required"`
}

type Body struct {
}

type Block struct {
	header *Header
}
