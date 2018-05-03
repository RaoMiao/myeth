package main

import (
	"mygostudy/myeth/utils"
	"os"

	"gopkg.in/urfave/cli.v1"
)

var (
	app = utils.NewApp()
)

func init() {
	app.Commands = []cli.Command{
		initCommand,
	}
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
