package p2p

import (
	"errors"
	"fmt"
	"io"
	"myeth/p2p/discover"
	"sort"
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
	in chan Msg //网络连接会把收到的网络包 放到chan变量里面 通知到eth里的handlemsg里面
	//通知变量
	closed <-chan struct{}
	wstart <-chan struct{}

	werr   chan<- error //网络连接写入的时候发生了错误
	offset uint64       //协议命令字的偏移量 每个service的msg code都是从0 开始的 所以要用偏移量 来区别
	w      MsgWriter    //网络conn 的接口 用来从protoRW里 发出去消息
}

func (rw *protoRW) WriteMsg(msg Msg) (err error) {
	if msg.Code >= rw.Length {
		return errors.New("超出了这个协议的命令字范围")
	}

	msg.Code += rw.offset
	select {
	case <-rw.wstart:
		//通过rlpx 将数据发出去了
		err = rw.w.WriteMsg(msg)
		rw.werr <- err
	case <-rw.closed:
		err = fmt.Errorf("shutting down")
	}
	return err
}

//从上层逻辑eth中调用进来 会一直阻塞住 知道有一个消息从protoRW的 in chan中传进来
func (rw *protoRW) ReadMsg() (Msg, error) {
	select {
	case msg := <-rw.in:
		//在上层逻辑处理之前 先进行msgcode 的偏移
		msg.Code -= rw.offset
		return msg, nil
	case <-rw.closed:
		return Msg{}, io.EOF
	}
}

//节点的连接结构
// Peer represents a connected remote node.
type Peer struct {
	//数据相关
	rw      *conn               //网络连接 通过tcp建立的网络连接
	running map[string]*protoRW //此Peer跑的协议

	wg       sync.WaitGroup
	protoErr chan error //协议层错误 收发包错误
	closed   chan struct{}
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

//创建 protoRW proto Reader Writer功能 主要的功能是对每个service的msgcode 做一个映射
func matchProtocols(protocols []Protocol, caps []Cap, rw MsgReadWriter) map[string]*protoRW {
	//将支持的协议数组进行排序
	sort.Sort(capsByNameAndVersion(caps))
	offset := baseProtocolLength
	result := make(map[string]*protoRW)

outer:
	//对本地协议和远端peer支持的协议进行逐个的比较
	for _, cap := range caps {
		for _, proto := range protocols {
			//协议 和 版本号 一致
			if proto.Name == cap.Name && proto.Version == cap.Version {
				//因为caps数组已经按照版本号排序了 所以如果找到了同名协议 肯定是之前有一个老的版本在map里了
				if old := result[cap.Name]; old != nil {
					//新的协议会覆盖老的 所以把老协议删掉
					offset -= old.Length
				}

				result[cap.Name] = &protoRW{Protocol: proto, offset: offset, in: make(chan Msg), w: rw}

				continue outer
			}
		}
	}

	return result
}

func newPeer(conn *conn, protocols []Protocol) *Peer {
	protomap := matchProtocols(protocols, conn.caps, conn)
	p := &Peer{
		rw:      conn,
		running: protomap,
	}
	return p
}

func (p *Peer) ID() discover.NodeID {
	return p.rw.id
}

//开启一个循环 读取网络消息
func (p *Peer) readLoop(errc chan<- error) {
	defer p.wg.Done()
	for {
		//rlpx 读取一个消息包
		msg, err := p.rw.ReadMsg()
		if err != nil {
			errc <- err
			return
		}

		if err = p.handle(msg); err != nil {
			errc <- err
			return
		}
	}
}

func (p *Peer) getProto(code uint64) (*protoRW, error) {
	for _, proto := range p.running {
		//判断msgcode 所落在的区间
		if code >= proto.offset && code < proto.offset+proto.Length {
			return proto, nil
		}
	}
	return nil, errors.New("fuck")
}

//tcp消息的第一层处理
func (p *Peer) handle(msg Msg) error {
	switch {
	case msg.Code == pingMsg:
		msg.Discard()
		go SendItems(p.rw, pongMsg)
	default:
		//eth子协议
		proto, err := p.getProto(msg.Code)
		if err != nil {
			return fmt.Errorf("msg code out of range: %v", msg.Code)
		}

		select {
		case proto.in <- msg:
			return nil
		case <-p.closed:
			return io.EOF
		}
	}
	return nil
}

//定时发送心跳包
func (p *Peer) pingLoop() {
	ping := time.NewTimer(pingInterval)
	defer p.wg.Done()
	defer ping.Stop()
	for {
		select {
		case <-ping.C:
			if err := SendItems(p.rw, pingMsg); err != nil {
				p.protoErr <- err
				return
			}
			ping.Reset(pingInterval)
		case <-p.closed:
			return
		}
	}
}

func (p *Peer) run() (err error) {
	var (
		readErr  = make(chan error, 1)
		writeErr = make(chan error, 1)
	)

	p.wg.Add(2)
	go p.readLoop(readErr)
	go p.pingLoop()
	p.startProtocols(writeErr)

loop:
	//启动一个循环 来等待连接结束
	for {
		select {

		case err = <-readErr:
			//读取部分发生错误
			break loop
		case err = <-p.protoErr:
			break loop
		}
	}

	return nil
}

func (p *Peer) startProtocols(writeErr chan<- error) {
	for _, proto := range p.running {
		proto.werr = writeErr

		//执行每一个协议的run函数 将他们启动起来
		go func() {
			err := proto.Run(p, proto)
		}()
	}
}
