package p2p

import (
	"crypto/ecdsa"
	"errors"
	"net"
	"sync"
	"time"

	"myeth/p2p/discover"
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

	// Static nodes are used as pre-configured connections which are always
	// maintained and re-connected on disconnects.
	StaticNodes []*discover.Node

	//节点名称
	// Name sets the node name of this server.
	// Use common.MakeName to create a name that follows existing conventions.
	Name string `toml:"-"`

	// dialTask 用来生成网络连接的dialer
	// If Dialer is set to a non-nil value, the given Dialer
	// is used to dial outbound peer connections.
	Dialer NodeDialer `toml:"-"`
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

	lock sync.Mutex // protects running

	running bool

	listener net.Listener
	//本节点的握手包
	ourHandshake *protoHandshake

	loopWG sync.WaitGroup // loop, listenLoop

	quit          chan struct{}
	posthandshake chan *conn
	addpeer       chan *conn
	delpeer       chan peerDrop
}

//用来描述断掉连接的结构体
type peerDrop struct {
	*Peer
	err error
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

	doEncHandshake(prv *ecdsa.PrivateKey, dialDest *discover.Node) (discover.NodeID, error)
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

func (srv *Server) newTransport(fd net.Conn) transport {
	return newRLPX(fd)
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

	//srv.newTransport = newRLPX

	srv.quit = make(chan struct{})
	srv.addpeer = make(chan *conn)
	srv.delpeer = make(chan peerDrop)
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

	if srv.Dialer == nil {
		srv.Dialer = TCPDialer{&net.Dialer{Timeout: 10 * time.Second}}
	}

	dialer := newDialState(srv.StaticNodes)
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
	go srv.run(dialer)
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

	//最大排队等待的链接个数
	tokens := 50

	//struct{} 用来做signal size为0
	slots := make(chan struct{}, tokens)
	for i := 0; i < tokens; i++ {
		//类型是 struct{} /调用struct{}{}是生成一个struct{}变量
		slots <- struct{}{}
	}

	for {
		//占用一个chan 然后开始等待一个链接的操作
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

		//以太坊在这里多做一层IP限制

		go func() {
			//链接建立 进行握手 同时释放掉 占用的chan
			go srv.SetupConn(fd, 0, nil)
			slots <- struct{}{}
		}()
	}
}

//握手的工作
func (srv *Server) SetupConn(fd net.Conn, flags connFlag, dialDest *discover.Node) error {
	c := &conn{fd: fd, transport: srv.newTransport(fd), flags: flags, cont: make(chan error)}
	err := srv.setupConn(c, flags, dialDest)
	if err != nil {
		c.close(err)
	}
	return err
}

func (srv *Server) setupConn(c *conn, flags connFlag, dialDest *discover.Node) error {
	// Prevent leftover pending conns from entering the handshake.
	srv.lock.Lock()
	running := srv.running
	srv.lock.Unlock()
	if !running {
		return nil
	}
	// Run the encryption handshake.
	var err error
	//NEED DO!!这里的c.id 和 dialDest里面的有不同吗
	if c.id, err = c.doEncHandshake(srv.PrivateKey, dialDest); err != nil {
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

//等待下一步操作的完成
func (srv *Server) checkpoint(c *conn, stage chan<- *conn) error {
	//将要做逻辑的节点连接 发送到p2pServer的Run loop中去做逻辑
	select {
	case stage <- c:
	case <-srv.quit:
		return nil
	}

	//在此等待p2pServer的Run loop中的链接 操作执行完毕 并返回对应错误
	select {
	case err := <-c.cont:
		return err
	case <-srv.quit:
		return nil
	}
}

const (
	maxActiveDialTasks = 16
)

//server主循环
func (srv *Server) run(dialer *dialstate) {
	defer srv.loopWG.Done()

	var (
		//当前连接到的节点map
		peers = make(map[discover.NodeID]*Peer)
		//任务执行完成后的通知chan列表
		taskdone = make(chan task, maxActiveDialTasks)
		//正在执行的task
		runningTasks []task
		//等待执行的task
		queuedTasks []task
	)

	//删除一个正在执行的任务
	delTask := func(t task) {
		for i := range runningTasks {
			if runningTasks[i] == t {
				runningTasks = append(runningTasks[:i], runningTasks[i+1:]...)
				break
			}
		}
	}

	//接收参数是要执行的任务列表
	startTasks := func(ts []task) (rest []task) {
		i := 0
		for ; len(runningTasks) < maxActiveDialTasks && i < len(ts); i++ {
			t := ts[i]
			//对于每个任务 开启一个新携程去执行
			go func() {
				t.Do(srv)
				taskdone <- t
			}()
			//将任务加入到执行列表中
			runningTasks = append(runningTasks, t)
		}
		//返回剩下未执行的task
		return ts[i:]
	}

	//任务调度函数
	scheduleTasks := func() {
		queuedTasks = startTasks(queuedTasks)
		//当前运行任务不够最大可运行数 创建更多的新任务
		if len(runningTasks) < maxActiveDialTasks {
			nt := dialer.newTasks()
			queuedTasks = append(queuedTasks, startTasks(nt)...)
		}
	}

running:
	for {
		//调度任务到运行队列
		scheduleTasks()

		select {
		case <-srv.quit:
			break running

		case t := <-taskdone:
			delTask(t)

		case c := <-srv.posthandshake:
			//第一阶段加密handshake操作完毕

			select {
			case c.cont <- srv.encHandshakeChecks(c):
			case <-srv.quit:
				break running
			}
		case c := <-srv.addpeer:
			//doProtoHandshake
			err := srv.protoHandshakeChecks(c)
			if err == nil {
				//握手完成 run peer开始
				// The handshakes are done and it passed all checks.
				p := newPeer(c, srv.Protocols)

				go srv.runPeer(p)
				peers[c.id] = p
			}

			select {
			case c.cont <- err:
			case <-srv.quit:
				break running
			}
		case pd := <-srv.delpeer:
			delete(peers, pd.ID())
		}

	}
}

func (srv *Server) protoHandshakeChecks(c *conn) error {
	//先检测协议是否能匹配
	if len(srv.Protocols) > 0 && countMatchingProtocols(srv.Protocols, c.caps) == 0 {
		return nil
	}

	return srv.encHandshakeChecks(c)
}

//这里面有
func (srv *Server) encHandshakeChecks(c *conn) error {
	return nil
	// switch {
	// case !c.is(trustedConn|staticDialedConn) && len(peers) >= srv.MaxPeers:
	// 	return DiscTooManyPeers
	// case !c.is(trustedConn) && c.is(inboundConn) && inboundCount >= srv.maxInboundConns():
	// 	return DiscTooManyPeers
	// case peers[c.id] != nil:
	// 	return DiscAlreadyConnected
	// case c.id == srv.Self().ID:
	// 	return DiscSelf
	// default:
	// 	return nil
	// }
}

//run peer 为每一个peer 开启一个 goroutine
func (srv *Server) runPeer(p *Peer) {
	err := p.run()

	srv.delpeer <- peerDrop{p, err}
}
