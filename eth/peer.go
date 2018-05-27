package eth

import (
	"myeth/p2p"
)

//对p2p.Peer的上层包装
type peer struct {
	*p2p.Peer
	protoRW p2p.MsgReadWriter //这里是protoRW结构
	version int
}

//rw 是 protoRW
func newPeer(version int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	return &peer{
		Peer:    p,
		protoRW: rw,
		version: version,
	}
}

// 发送一个头部数组 给一个 节点
// func (p *peer) SendBlockHeaders(headers []*types.Header) error {
// 	return p2p.Send(p.protoRW, BlockHeadersMsg, headers)
// }
