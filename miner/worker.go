package miner

import (
	"math/big"
	"myeth/common"
	"myeth/core"
	"myeth/core/types"
	"time"
)

type Work struct {
	Block    *types.Block
	header   *types.Header
	txs      []*types.Transactions
	receipts []*types.Receipt

	createdAt time.Time
}

type worker struct {
	singer types.Signer

	coinbase common.Address

	chain *core.BlockChain

	engine consensus.Engine

	eth Backend

	current *Work
}

func newWorker() *worker {
	worker := &worker{}
	return worker
}

func (self *worker) commitNewWork() {
	tstart := time.Now()
	parent := self.chain.CurrentBlock()

	tstamp := tstart.Unix()
	//这句话什么意思
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}

	num := parent.Number()
	//构建新区快的头部
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number: num.Add(num, common.Big1),
		GasLimit: 0，		//这数据干嘛的
		Extra: self.extra,
		Time: big.NewInt(tstamp),
	}
	header.Coinbase = self.coinbase
	if err := self.engine.Prepare(self.chain, header); err != nil{
		return;
	}

	work := self.current

	pending, err := self.eth.TxPool().Pending()
	if err != nil{
		return
	}

	txs := types.NewTransactionsByPriceAndNonce(self.current.signer, pending)

}

func (env *Work) commitTransactions(txs *types.TransactionsByPriceAndNonce, bc *core.BlockChain, coinbase common.Address){
	
}