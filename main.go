package main

import (
	"jstarpl/jpm/client"
	"jstarpl/jpm/service"
	"log"
	"runtime"

	"github.com/alecthomas/kong"
	"github.com/ffred/guitocons"
)

type CLI struct {
	Service service.Service `cmd:"" help:"Manage the JPM service."`

	Ps      client.Ps      `cmd:"" help:"List running processes." aliases:"list,ls"`
	Start   client.Start   `cmd:"" help:"Start a new process." aliases:"add"`
	Stop    client.Stop    `cmd:"" help:"Stop specified process"`
	Restart client.Restart `cmd:"" help:"Restart an existing process"`
	Delete  client.Delete  `cmd:"" help:"Delete specified process (implies 'stop')" aliases:"del,rm"`
}

func main() {
	// this should probably actually use https://stackoverflow.com/questions/35428963/what-is-the-right-approach-to-encapsulate-platform-specific-code-in-go
	if runtime.GOOS == "windows" {
		guitocons.Guitocons()
	}

	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "service start":
		service.StartService(&cli.Service)
	case "service stop":
		client.RequestStopService()
	case "ps":
		client.ListProcesses(&cli.Ps)
	case "start <args>":
		client.StartProcess(&cli.Start)
	case "stop <id>":
		client.StopProcess(&cli.Stop)
	case "delete <id>":
		client.DeleteProcess(&cli.Delete)
	default:
		log.Default().Printf("Unknown command %s", ctx.Command())
		panic(ctx.Error)
	}
}
