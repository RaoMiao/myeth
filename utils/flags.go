package utils

import (
	cli "gopkg.in/urfave/cli.v1"
)

func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = "myeth"
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = "1.0"

	app.Usage = "no usage"
	return app
}
