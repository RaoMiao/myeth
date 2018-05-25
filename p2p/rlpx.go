package p2p

import (
	"crypto/ecdsa"
	"io"
	"myeth/p2p/discover"
	"net"
	"time"

	"github.com/seeleteam/go-seele/crypto/ecies"
)

const (
	// devp2p message codes
	handshakeMsg = 0x00
	discMsg      = 0x01
	pingMsg      = 0x02
	pongMsg      = 0x03
	getPeersMsg  = 0x04
	peersMsg     = 0x05
)

const (
	handshakeTimeout = 5 * time.Second
)

//rlpx 协议
type rlpx struct {
	fd net.Conn
}

func newRLPX(fd net.Conn) transport {
	fd.SetDeadline(time.Now().Add(handshakeTimeout))
	return &rlpx{fd: fd}
}

//rlpx 实现 transport协议
func (t *rlpx) ReadMsg() (Msg, error) {

}

func (t *rlpx) WriteMsg(msg Msg) error {

}

func (t *rlpx) close(err error) {

}

//p2p 双方都放松自己的握手包给对方 同时等待对方的握手包
//发送我们的握手包 给他们 接收他们的握手包
func (t *rlpx) doProtoHandshake(our *protoHandshake) (their *protoHandshake, err error) {
	werr := make(chan error, 1)
	go func() {
		//发送一个握手包
		werr <- Send(t.fd, handshakeMsg, our)
	}()
	if their, err = readProtocolHandshake(t.fd, our); err != nil {
		<-werr
		return nil, err
	}
	if err := <-werr; err != nil {
		return nil, nil
	}

	return their, nil
}

func readProtocolHandshake(rw MsgReader, our *protoHandshake) (*protoHandshake, error) {
	//ReadMsg 这里应该会堵塞住
	msg, err := rw.ReadMsg()
	if err != nil {
		return nil, err
	}

	if msg.Code != handshakeMsg {
		return nil, nil
	}

	var hs protoHandshake
	//将网络数据 解密成 握手结构体
	return &hs, nil
}

// doEncHandshake runs the protocol handshake using authenticated
// messages. the protocol handshake is the first authenticated message
// and also verifies whether the encryption handshake 'worked' and the
// remote side actually provided the right public key.
func (t *rlpx) doEncHandshake(prv *ecdsa.PrivateKey, dest *discover.Node) (discover.NodeID, error) {
	if dest == nil {
		//服务器端接收
		receiverEncHandshake(t.fd, prv)
	} else {
		//客户端主动连接
		initiatorEncHandshake(t.fd, prv, dest.ID)
	}

	return dest.ID, nil
}

type encHandshake struct {
	initiator bool
	destID    discover.NodeID

	remotePub *ecies.PublicKey // remote-pubk
}

//处理AuthMsg消息
func (h *encHandshake) handleAuthMsg(msg *authMsgV4, prv *ecdsa.PrivateKey) {

}

// RLPx v4 handshake auth (defined in EIP-8).
type authMsgV4 struct {
	Version uint
}

type authRespV4 struct {
	Version uint
}

func (h *encHandshake) makeAuthMsg(prv *ecdsa.PrivateKey) *authMsgV4 {
	msg := new(authMsgV4)
	msg.Version = 4
	return msg
}

func (h *encHandshake) makeAuthResp() (msg *authRespV4) {
	msg = new(authRespV4)
	msg.Version = 4
	return msg
}

//prv 是当前节点的私钥
func initiatorEncHandshake(conn io.ReadWriter, prv *ecdsa.PrivateKey, remoteID discover.NodeID) {
	h := &encHandshake{initiator: true, destID: remoteID}
	authMsg := h.makeAuthMsg(prv)

	if _, err := conn.Write(authMsg); err != nil {
		return
	}

	authRespV4 := new(authRespV4)
	authRespPacket, err := readHandshakeMsg(1024, prv, conn)
	if err != nil {

	}

}

func readHandshakeMsg(plainSize int, prv *ecdsa.PrivateKey, r io.Reader) ([]byte, error) {
	buf := make([]byte, plainSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return buf, err
	}

	//一大段解密
	return buf, nil
}

//服务端接收authMsg请求
func receiverEncHandshake(conn io.ReadWriter, prv *ecdsa.PrivateKey) {
	authMsg := new(authMsgV4)
	authPacket, err := readHandshakeMsg(1024, prv, conn)
	if err != nil {
		return
	}

	h := new(encHandshake)
	h.handleAuthMsg(authMsg, prv)

	authRespMsg := h.makeAuthResp()

	if _, err = conn.Write(authRespMsg); err != nil {
		return
	}
}
