package eth

import (
	"myeth/core"
	"myeth/node"
	"myeth/p2p"

	"myeth/ethdb"

	"github.com/ethereum/go-ethereum/params"
)

type Ethereum struct {
	protocolManager *ProtocolManager
}

// New creates a new Ethereum object (including the
// initialisation of the common Ethereum object)
func New(ctx *node.ServiceContext) (*Ethereum, error) {

	chainDb, err := CreateDB(ctx, "chaindata")
	if err != nil {
		return nil, err
	}

	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}

	eth := Ethereum{}

	if eth.protocolManager, err = NewProtocolManager(); err != nil {
		return nil, err
	}

	return &eth, nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Ethereum) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

func (s *Ethereum) Start(server *p2p.Server) error {
	return nil
}

func (s *Ethereum) Stop() error {
	return nil
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, name string) (ethdb.Database, error) {
	db, err := ctx.OpenDatabase(name, 1, 1)
	if err != nil {
		return nil, err
	}
	// if db, ok := db.(*ethdb.LDBDatabase); ok {

	// }
	return db, nil
}
