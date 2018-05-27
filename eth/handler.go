package eth

import "myeth/p2p"

// Official short name of the protocol used during capability negotiation.

type ProtocolManager struct {
	SubProtocols []p2p.Protocol

	newPeerCh chan *peer
	quitSync  chan struct{}
}

// NewProtocolManager returns a new Ethereum sub protocol manager. The Ethereum sub protocol manages peers capable
// with the Ethereum network.
func NewProtocolManager() (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{}

	//支持几套版本的协议 来创建几个protocol
	manager.SubProtocols = make([]p2p.Protocol, 0, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		version := version

		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  ProtocolLengths[i],
			//协议运行主函数
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				//创建一个上层peer
				peer := newPeer(int(version), p, rw)
				return manager.handle(peer)
			},
		})
	}
	return manager, nil
}

//每个p2p peer的生命周期函数 当退出时 peer就断开了
func (pm *ProtocolManager) handle(p *peer) error {

	//要执行一下 以太坊的握手协议 主要判断测试网 版本 创世块

	//同步本节点现在的交易池交易 同步给 这个节点

	//主循环 处理消息
	for {
		if err := pm.handleMsg(p); err != nil {
			return err
		}
	}

	return nil
}

//handleMsg 处理远端节点发来的入站消息
func (pm *ProtocolManager) handleMsg(p *peer) error {
	//protoRW 等待 in chan 被写入一个Msg结构
	msg, err := p.protoRW.ReadMsg()
	if err != nil {
		return err
	}

	//不知道这有什么作用
	//defer msg.Discard()
	//针对不同的msg code做不同的处理
	switch {
	case msg.Code == GetBlockHeadersMsg:

	}
	return nil
}
