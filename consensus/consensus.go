package consensus

import (
	"myeth/core/types"
)

type ChainReader interface {
	CurrentHeader() *types.Header

	GetHeaderByNumber(number uint64) *types.Header
}

type Engine interface {
	//TODO: comment
	Prepare(chain ChainReader, header *types.Header) error
}
