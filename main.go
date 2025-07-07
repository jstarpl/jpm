package main

import (
	"jstarpl/jpm/client"
	"jstarpl/jpm/service"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Service service.Service `cmd:"" help:"Manage the JPM service."`

	Ps    client.Ps    `cmd:"" help:"List running processes."`
	Start client.Start `cmd:"" help:"Start a new process."`
	Stop  client.Stop  `cmd:"" help:"Stop specified process"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "service start":
		service.StartService(&cli.Service)
	case "ps":
		client.ListProcesses(&cli.Ps)
	case "start":
		client.StartProcess(&cli.Start)
	default:
		panic(ctx.Error)
	}
}
