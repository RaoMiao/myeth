// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// geth is the official command-line client for Ethereum.
package main

import (
	"runtime"

	"myeth/eth"
	"myeth/node"
	"myeth/utils"
)

func main() {
	//设置携程使用数量
	runtime.GOMAXPROCS(runtime.NumCPU())
	geth()
}

type gethConfig struct {
	Eth  eth.Config
	Node node.Config
}

func makeConfigNode() *node.Node {

	stack, err := node.New()
	if err != nil {
		//utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	return stack
}

func makeFullNode() *node.Node {
	stack := makeConfigNode()

	//NEED DO!! 这里要高
	utils.RegisterEthService(stack)

	return stack
}

// geth is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func geth() error {
	//创建一个全节点
	node := makeFullNode()
	//开始节点工作
	startNode(node)
	//等待退出命令
	node.Wait()
	return nil
}

func startNode(stack *node.Node) {
	stack.Start()
}
