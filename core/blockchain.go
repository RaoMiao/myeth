package core

import (
	"myeth/core/types"
	"myeth/ethdb"
)

type BlockChain struct {
	db ethdb.Database

	hc           *HeaderChain //区块头链
	genesisBlock *types.Block //创世块
}

//创建一个区块链结构
func NewBlockChain(db ethdb.Database) (*BlockChain, error) {
	bc := &BlockChain{
		db: db,
	}

	var err error
	bc.hc, err = NewHeaderChain(db)
	if err != nil {
		return nil, err
	}

	return bc, nil
}
