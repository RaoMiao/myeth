package types

import (
	"math/big"
	"myeth/common"
)

// Header represents a block header in the Ethereum blockchain.
// 以太坊区块头
type Header struct {
	ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash    `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address `json:"miner"            gencodec:"required"`
	Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	//Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
	Difficulty *big.Int    `json:"difficulty"       gencodec:"required"`
	Number     *big.Int    `json:"number"           gencodec:"required"`
	GasLimit   uint64      `json:"gasLimit"         gencodec:"required"`
	GasUsed    uint64      `json:"gasUsed"          gencodec:"required"`
	Time       *big.Int    `json:"timestamp"        gencodec:"required"`
	Extra      []byte      `json:"extraData"        gencodec:"required"`
	MixDigest  common.Hash `json:"mixHash"          gencodec:"required"`
	//Nonce       BlockNonce     `json:"nonce"            gencodec:"required"`
}

// Block represents an entire block in the Ethereum blockchain.
// 以太坊区块结构体
type Block struct {
	header       *Header
	uncles       []*Header
	transactions Transactions
}
