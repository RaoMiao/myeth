package core

import (
	"myeth/common"
	"myeth/core/types"
)

//从网络中接收进入 被包含进区块链离开
type TxPool struct {
}

//遍历交易 按地址分类
func (pool *TxPool) Pending() (map[common.Address]types.Transactions, error) {
	pending := make(map[common.Address]types.Transactions)

	return pending, nil
}
