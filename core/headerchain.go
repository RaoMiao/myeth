package core

import (
	"myeth/core/types"
	"myeth/ethdb"
)

type HeaderChain struct {
	chainDb       ethdb.Database
	genesisHeader *types.Header
}

func NewHeaderChain(chainDb ethdb.Database) (*HeaderChain, error) {
	hc := &HeaderChain{
		chainDb: chainDb,
	}

	return hc, nil
}
