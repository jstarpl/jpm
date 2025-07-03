package main

import (
	"jstarpl/jpm/client"
	"jstarpl/jpm/service"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Service service.Service `cmd:"" help:"Manage the JPM service."`

	Ps struct{} `cmd:"" help:"List running processes."`

	Start struct {
		Name []string `name:"name" help:"Name of the process"`
		Args []string `arg:""`
	} `cmd:"" help:"Start a new process."`

	Stop struct{} `cmd:"" help:"Stop specified process"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "service start":
		service.StartService(&cli.Service)
	case "ps":
		client.ListProcesses()
	default:
		panic(ctx.Error)
	}
}
