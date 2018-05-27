package node

import (
	"reflect"
	"sync"

	"myeth/p2p"
)

//Node 包含 p2pServer 和 依赖p2pServer的上层 Serivce对象
type Node struct {
	config *Config
	//账户相关
	//accman   *accounts.Manager

	serverConfig p2p.Config
	server       *p2p.Server // Currently running P2P networking layer

	serviceFuncs []ServiceConstructor     // Service constructors (in dependency order)
	services     map[reflect.Type]Service // Currently running services

	stop chan struct{} // Channel to wait for termination notifications
	lock sync.RWMutex
}

// New creates a new P2P node, ready for protocol registration.
func New(conf *Config) (*Node, error) {
	confCopy := *conf

	return &Node{
		config: &confCopy,
	}, nil
}

func (n *Node) Wait() {
	stop := n.stop
	<-stop
}

// Register injects a new service into the node's stack. The service created by
// the passed constructor must be unique in its type with regard to sibling ones.
func (n *Node) Register(constructor ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.server != nil {
		return nil
	}
	n.serviceFuncs = append(n.serviceFuncs, constructor)
	return nil
}

//Node启动流程
// 1. 创建p2pServer
// 2. 创建ethereum对象 包含所有的以太坊逻辑处理
// 3. 配置以太坊协议到p2pServer的config里面去
// 4. 启动p2pServer
// 5. 启动以太坊逻辑
func (n *Node) Start() error {

	//初始化P2P Server
	//n.serverConfig = n.config.P2P
	n.config.DataDir = "./geth/"
	n.serverConfig.PrivateKey = n.config.NodeKey()
	if n.serverConfig.StaticNodes == nil {
		n.serverConfig.StaticNodes = n.config.StaticNodes()
	}
	n.serverConfig.ListenAddr = n.config.P2P.ListenAddr

	running := &p2p.Server{Config: n.serverConfig}

	// Otherwise copy and specialize the P2P configuration
	services := make(map[reflect.Type]Service)
	for _, constructor := range n.serviceFuncs {
		// Create a new context for the particular service
		ctx := &ServiceContext{
			config: n.config,
		}

		// Construct and save the service
		service, err := constructor(ctx)
		if err != nil {
			return err
		}
		kind := reflect.TypeOf(service)
		services[kind] = service
	}

	// 本节点所支持的协议 是所有服务 共同提供的
	// Gather the protocols and start the freshly assembled P2P server
	for _, service := range services {
		running.Protocols = append(running.Protocols, service.Protocols()...)
	}

	//开始P2P Server
	if err := running.Start(); err != nil {
		return err
	}

	//开启各大服务器
	// Start each of the services
	started := []reflect.Type{}
	for kind, service := range services {
		// Start the next service, stopping all previous upon failure
		if err := service.Start(running); err != nil {
			for _, kind := range started {
				services[kind].Stop()
			}
			running.Stop()

			return err
		}
		// Mark the service started for potential cleanup
		started = append(started, kind)
	}

	n.server = running
	n.services = services
	n.stop = make(chan struct{})

	return nil
}
