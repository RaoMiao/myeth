package utils

import (
	"myeth/eth"
	"myeth/node"
)

//创建一个全节点
func RegisterEthService(stack *node.Node) {
	var err error

	err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		fullNode, err := eth.New(ctx)
		return fullNode, err
	})

	if err != nil {
		return
	}
}
