package miner

import (
	"myeth/common"
	"myeth/core"
	"myeth/ethdb"
)

//挖矿需要的数据支持
type Backend interface {
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
	ChainDB() ethdb.Database
}

//Miner creates blocks and searches for proof-of-work values
type Miner struct {

	//矿工给自己的地址 加币
	coinbase common.Address
}
