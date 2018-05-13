package p2p

import (
	"crypto/ecdsa"
	"errors"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/p2p/discover"
)

// Config holds Server options.
type Config struct {
	//需要弄明白这个私钥的来源
	// This field must be set to a valid secp256k1 private key.
	PrivateKey *ecdsa.PrivateKey `toml:"-"`

	// 这个P2P节点所支持的协议
	// Protocols should contain the protocols supported
	// by the server. Matching protocols are launched for
	// each peer.
	Protocols []Protocol `toml:"-"`

	// If ListenAddr is set to a non-nil address, the server
	// will listen for incoming connections.
	//
	// If the port is zero, the operating system will pick a port. The
	// ListenAddr field will be updated with the actual address when
	// the server is started.
	ListenAddr string

	//节点名称
	// Name sets the node name of this server.
	// Use common.MakeName to create a name that follows existing conventions.
	Name string `toml:"-"`
}

type connFlag int

// conn wraps a network connection with information gathered
// during the two handshakes.
type conn struct {
	fd net.Conn
	transport
	flags connFlag
	cont  chan error      // The run loop uses cont to signal errors to SetupConn.
	id    discover.NodeID // valid after the encryption handshake
	caps  []Cap           // valid after the protocol handshake
	name  string          // valid after the protocol handshake
}

// Server manages all peer connections.
// p2pServer 管理所有连接
type Server struct {
	// Config fields may not be modified while the server is running.
	Config

	newTransport func(net.Conn) transport

	lock sync.Mutex // protects running

	running bool

	listener net.Listener
	//本节点的握手包
	ourHandshake *protoHandshake

	loopWG sync.WaitGroup // loop, listenLoop

	quit          chan struct{}
	posthandshake chan *conn
	addpeer       chan *conn
}

// sharedUDPConn implements a shared connection. Write sends messages to the underlying connection while read returns
// messages that were found unprocessable and sent to the unhandled channel by the primary listener.
type sharedUDPConn struct {
	*net.UDPConn
	//unhandled chan discover.ReadPacket
}

//transport 是什么作用
type transport interface {
	// The two handshakes.

	doEncHandshake(prv *ecdsa.PrivateKey) error
	doProtoHandshake(our *protoHandshake) (*protoHandshake, error)

	// The MsgReadWriter can only be used after the encryption
	// handshake has completed. The code uses conn.id to track this
	// by setting it to a non-nil value after the encryption handshake.
	MsgReadWriter
	// transports must provide Close because we use MsgPipe in some of
	// the tests. Closing the actual network connection doesn't do
	// anything in those tests because NsgPipe doesn't use it.
	close(err error)
}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (srv *Server) Start() (err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true

	srv.newTransport = newRLPX

	srv.quit = make(chan struct{})
	srv.addpeer = make(chan *conn)
	srv.posthandshake = make(chan *conn)

	//discovery 功能先屏蔽掉
	// var (
	// 	conn     *net.UDPConn
	// 	realaddr *net.UDPAddr
	// )

	// addr, err := net.ResolveUDPAddr("udp", ":30303")
	// if err != nil {
	// 	return err
	// }
	// conn, err = net.ListenUDP("udp", addr)
	// if err != nil {
	// 	return err
	// }
	// realaddr = conn.LocalAddr().(*net.UDPAddr)

	// cfg := discover.Config{
	// 	PrivateKey:   srv.PrivateKey,
	// 	AnnounceAddr: realaddr,
	// 	NodeDBPath:   srv.NodeDatabase,
	// 	NetRestrict:  srv.NetRestrict,
	// 	Bootnodes:    srv.BootstrapNodes,
	// 	Unhandled:    unhandled,
	// }
	// ntab, err := discover.ListenUDP(conn, cfg)
	// if err != nil {
	// 	return err
	// }
	// srv.ntab = ntab

	// handshake
	// 本节点的握手包
	srv.ourHandshake = &protoHandshake{Version: baseProtocolVersion, Name: srv.Name}
	for _, p := range srv.Protocols {
		srv.ourHandshake.Caps = append(srv.ourHandshake.Caps, p.cap())
	}

	if err := srv.startListening(); err != nil {
		return err
	}

	srv.loopWG.Add(1)
	go srv.run()
	srv.running = true
	return nil
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *Server) Stop() {
}

//开启TCP 监听别的peer消息
func (srv *Server) startListening() error {
	// Launch the TCP listener.
	listener, err := net.Listen("tcp", ":30303")
	if err != nil {
		return err
	}
	laddr := listener.Addr().(*net.TCPAddr)
	srv.ListenAddr = laddr.String()
	srv.listener = listener
	srv.loopWG.Add(1)
	go srv.listenLoop()

	return nil
}

//这是什么东西 有待验证
type tempError interface {
	Temporary() bool
}

// listenLoop runs in its own goroutine and accepts
// inbound connections.
// 监听NewPeer的 routine
func (srv *Server) listenLoop() {
	//在外面给waitgroup +1 进到goroutine 先给waitgroup 减1
	defer srv.loopWG.Done()

	tokens := 50

	slots := make(chan struct{}, tokens)
	for i := 0; i < tokens; i++ {
		slots <- struct{}{}
	}

	for {
		<-slots

		var (
			fd  net.Conn //tcp 链接
			err error
		)

		for {
			fd, err = srv.listener.Accept()
			if tempErr, ok := err.(tempError); ok && tempErr.Temporary() {
				continue
			}
			//接入了一个无错链接
			break
		}

		go func() {
			go srv.SetupConn(fd, 0)
			slots <- struct{}{}
		}()
	}
}

func (srv *Server) SetupConn(fd net.Conn, flags connFlag) error {
	c := &conn{fd: fd, transport: srv.newTransport(fd), flags: flags, cont: make(chan error)}
	err := srv.setupConn(c, flags)
	if err != nil {
		c.close(err)
	}
	return err
}

func (srv *Server) setupConn(c *conn, flags connFlag) error {
	// Prevent leftover pending conns from entering the handshake.
	srv.lock.Lock()
	running := srv.running
	srv.lock.Unlock()
	if !running {
		return nil
	}
	// Run the encryption handshake.
	var err error
	if err = c.doEncHandshake(srv.PrivateKey); err != nil {
		//srv.log.Trace("Failed RLPx handshake", "addr", c.fd.RemoteAddr(), "conn", c.flags, "err", err)
		return err
	}

	err = srv.checkpoint(c, srv.posthandshake)
	if err != nil {
		return err
	}
	// Run the protocol handshake
	phs, err := c.doProtoHandshake(srv.ourHandshake)
	if err != nil {
		return err
	}

	c.caps, c.name = phs.Caps, phs.Name
	err = srv.checkpoint(c, srv.addpeer)
	if err != nil {
		return err
	}
	// If the checks completed successfully, runPeer has now been
	// launched by run.
	return nil
}

//这个函数的意义 需要研究
func (srv *Server) checkpoint(c *conn, stage chan<- *conn) error {
	select {
	case stage <- c:
	case <-srv.quit:
		return nil
	}

	select {
	case err := <-c.cont:
		return err
	case <-srv.quit:
		return nil
	}
}

//server主循环
func (srv *Server) run() {
	defer srv.loopWG.Done()

running:
	for {
		select {
		case <-srv.quit:
			break running
		}
	}
}
