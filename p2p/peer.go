package p2p

import (
	"myeth/p2p/discover"
	"sync"
	"time"
)

const (
	baseProtocolVersion    = 5
	baseProtocolLength     = uint64(16)
	baseProtocolMaxMsgSize = 2 * 1024

	snappyProtocolVersion = 5

	pingInterval = 15 * time.Second
)

//proto Reader Writer
//相关相关的读写
type protoRW struct {
	Protocol
	in     chan Msg        // receices read messages
	closed <-chan struct{} // receives when peer is shutting down
	wstart <-chan struct{} // receives when write may start
	werr   chan<- error    // for write results
	offset uint64
	w      MsgWriter
}

//节点的连接结构
// Peer represents a connected remote node.
type Peer struct {
	//数据相关
	rw      *conn               //网络连接 在server中定义
	running map[string]*protoRW //mei

	//结构相关
	wg       sync.WaitGroup
	protoErr chan error
	closed   chan struct{}
	disc     chan DiscReason
}

//握手协议
// protoHandshake is the RLP structure of the protocol handshake.
type protoHandshake struct {
	Version    uint64
	Name       string
	Caps       []Cap
	ListenPort uint64
}

//统计两个协议数组能够对上的协议
func countMatchingProtocols(protocols []Protocol, caps []Cap) int {
	n := 0
	for _, cap := range caps {
		for _, proto := range protocols {
			if proto.Name == cap.Name && proto.Version == cap.Version {
				n++
			}
		}
	}
	n++
	return n
}

func matchProtocols(protocols []Protocol, caps []Cap, rw MsgReadWriter) map[string]*protoRW {

}

func newPeer(conn *conn, protocols []Protocol) *Peer {
	protomap := matchProtocols(protocols, conn.caps, conn)
	p := &Peer{}
	return p
}

func (p *Peer) ID() discover.NodeID {
	return p.rw.id
}

func (p *Peer) run() (err error) {
	return nil
}
