package core

import (
	"math/big"
	"myeth/common"
	"myeth/core/rawdb"
	"myeth/core/state"
	"myeth/core/types"
	"myeth/ethdb"

	"myeth/params"
)

// 创世区块里面的账号结构
// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

type Genesis struct {
	Config     *params.ChainConfig `json:"config"`
	Nonce      uint64              `json:"nonce"`
	Timestamp  uint64              `json:"timestamp"`
	ExtraData  []byte              `json:"extraData"`
	GasLimit   uint64              `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int            `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash         `json:"mixHash"`
	Coinbase   common.Address      `json:"coinbase"`
	Alloc      GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

// 以太坊主网的默认创世块
func DefaultGenesisBlock() *Genesis {
	return &Genesis{
	// Config:     params.MainnetChainConfig,
	// Nonce:      66,
	// ExtraData:  hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
	// GasLimit:   5000,
	// Difficulty: big.NewInt(17179869184),
	// Alloc:      decodePrealloc(mainnetAllocData),
	}
}

func (g *Genesis) ToBlock(db ethdb.Database) *types.Block {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		//statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		// for key, value := range account.Storage {
		// 	statedb.SetState(addr, key, value)
		// }
	}
	root := statedb.IntermediateRoot(false)
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)
}

//将创世块提交进leveldb
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {

}

func SetupGenesisBlock(db ethdb.Database, genesis *Genesis) (common.Hash, error) {
	//从db里面查创世块的hash
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		//没有找到创世块 使用默认创世块
		genesis := DefaultGenesisBlock()
		block, err := genesis.Commit(db)
	}
}
