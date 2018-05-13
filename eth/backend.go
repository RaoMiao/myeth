package eth

import (
	"myeth/node"
	"myeth/p2p"
)

type Ethereum struct {
	protocolManager *ProtocolManager
}

// New creates a new Ethereum object (including the
// initialisation of the common Ethereum object)
func New(ctx *node.ServiceContext) (*Ethereum, error) {
	eth := Ethereum{}

	var err error
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
