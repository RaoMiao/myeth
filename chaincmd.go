package main

import (
	"math/big"
	"mygostudy/myeth/common"
	"mygostudy/myeth/core"
	"mygostudy/myeth/ethdb"
	"mygostudy/myeth/params"

	"gopkg.in/urfave/cli.v1"
)

var (
	initCommand = cli.Command{
		Action: initGenesis,
		Name:   "init",
	}
)

// OpenDatabase opens an existing database with the given name (or creates one if no
// previous can be found) from within the node's instance directory. If the node is
// ephemeral, a memory database is returned.
func OpenDatabase() (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(".\\chaindb", 0, 0)
}

//创建创世区块
func initGenesis(ctx *cli.Context) error {
	genesis := new(core.Genesis)
	genesis.Config = new(params.ChainConfig)
	genesis.Config.ChainId = big.NewInt(0)
	genesis.Config.HomesteadBlock = big.NewInt(0)
	genesis.Config.EIP155Block = big.NewInt(0)
	genesis.Config.EIP158Block = big.NewInt(0)

	genesis.Nonce = 0x0000000000000042
	genesis.Timestamp = 0x0
	genesis.ParentHash = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	genesis.GasLimit = 0x80000000
	genesis.Difficulty = big.NewInt(0x1)
	genesis.Mixhash = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	genesis.Coinbase = common.HexToAddress("0x3333333333333333333333333333333333333333")

	chaindb, _ := OpenDatabase()
	core.SetupGenesisBlock(chaindb, genesis)

	return nil
}
