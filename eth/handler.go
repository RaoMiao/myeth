package eth

import "myeth/p2p"

// Official short name of the protocol used during capability negotiation.

type ProtocolManager struct {
	SubProtocols []p2p.Protocol
}

// NewProtocolManager returns a new Ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the Ethereum network.
func NewProtocolManager() (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{}
	return manager, nil
}
