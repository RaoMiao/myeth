package p2p

import (
	"errors"
	"myeth/p2p/discover"
	"net"
)

type task interface {
	Do(*Server)
}

type dialTask struct {
	//连接状态flag
	flags connFlag

	//要连接的目标节点信息
	dest *discover.Node
}

//dialTask 对目标 node发起tcp连接
func (t *dialTask) Do(srv *Server) {
	t.dial(srv, t.dest)
}

//dial 先进行TCP 连接 再进行握手检测
func (t *dialTask) dial(srv *Server, dest *discover.Node) error {
	//使用server的dialer 生成tcp连接
	fd, err := srv.Dialer.Dial(dest)
	if err != nil {
		return err
	}
	return srv.SetupConn(fd, t.flags, dest)
}

type dialstate struct {

	//当前的节点dialing状态map
	dialing map[discover.NodeID]connFlag

	//静态节点的连接任务map
	static map[discover.NodeID]*dialTask
}

func newDialState(static []*discover.Node) *dialstate {
	s := &dialstate{}
	for _, n := range static {
		s.addStatic(n)
	}
	return s
}

func (s *dialstate) addStatic(n *discover.Node) {
	s.static[n.ID] = &dialTask{dest: n}
}

var (
	errSelf             = errors.New("is self")
	errAlreadyDialing   = errors.New("already dialing")
	errAlreadyConnected = errors.New("already connected")
	errRecentlyDialed   = errors.New("recently dialed")
	errNotWhitelisted   = errors.New("not contained in netrestrict whitelist")
)

//节点连接状态检查 用error来表示连接状态
func (s *dialstate) checkDial(n *discover.Node) error {
	//从map中 取值 vluae, ok := map[key]
	_, dialing := s.dialing[n.ID]
	switch {
	case dialing:
		return errAlreadyDialing
	}
	return nil
}

//创建新的任务
func (s *dialstate) newTasks() []task {
	var newtasks []task

	//如果静态节点没有连接 创建任务连接
	for id, t := range s.static {
		err := s.checkDial(t.dest)
		switch err {
		case nil:
			//没有加入dialing map中 说明还没有尝试连接过
			s.dialing[id] = t.flags
			newtasks = append(newtasks, t)
		}
	}

	return newtasks
}

//对一个节点发起连接 返回一个网络连接
type NodeDialer interface {
	Dial(*discover.Node) (net.Conn, error)
}

//以太坊使用TCP对发现的节点进行连接
type TCPDialer struct {
	*net.Dialer
}

func (t TCPDialer) Dial(dest *discover.Node) (net.Conn, error) {
	addr := net.TCPAddr{IP: dest.IP, Port: int(dest.TCP)}
	return t.Dialer.Dial("tcp", addr.String())
}
