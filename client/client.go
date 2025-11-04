package client

import (
	"fmt"
	"jstarpl/jpm/api"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Ps struct{}

type Start struct {
	Name string   `name:"name" help:"Name of the process"`
	Args []string `arg:""`
}

type Stop struct {
	Id string `arg:""`
}

type Restart struct {
	Id string `arg:""`
}

type Delete struct {
	Id string `arg:""`
}

func ListProcesses(cli *Ps) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}

	req := &api.RequestListProcessesParams{}
	bReq, err := api.NewRequest(1, req)
	if err != nil {
		panic("Could not build request")
	}

	err = client.WriteMsg(bReq)
	if err != nil {
		log.Fatalf("Unknown writing to server: %v", err)
	}

	data, err := client.ReadMsg()

	if err != nil {
		log.Fatalf("Unknown error reading from server: %v", err)
	}

	res, err := api.UnmarshalResponse(data)

	if err != nil {
		log.Fatalf("Could not decode response: %v", err)
	}

	if res.Error != nil {
		log.Fatalf("Error while doing. %d %s", res.Error.Code, res.Error.Message)
	}

	if res.Result == nil || res.Result.ProcessList == nil {
		log.Fatalf("Invalid response: %v", res)
	}

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"ID", "Name", "Command", "Status", "Uptime", "Args"})
	for _, process := range *res.Result.ProcessList {
		tw.AppendRow(table.Row{process.Id, process.Name, process.Exec, process.Status, time.Duration(process.Uptime) * time.Millisecond, strings.Join(process.Arg, " ")})
	}
	tw.SetStyle(table.StyleRounded)
	fmt.Println(tw.Render())

	client.Close()
}

func StartProcess(cli *Start) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}
	defer client.Close()

	pwd, err := os.Getwd()
	if err != nil {
		pwd = os.TempDir()
	}

	req := &api.RequestStartProcessParams{
		Name: cli.Name,
		Exec: cli.Args[0],
		Arg:  cli.Args[1:],
		Env:  os.Environ(),
		Dir:  pwd,
	}
	SendRequest(client, 1, req)
	res, _ := ReadResponse(client)

	if res.Result != nil && res.Result.Success != nil {
		fmt.Printf("Process started %s", *res.Result.ProcessId)
	}

}

func StopProcess(cli *Stop) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}
	defer client.Close()

	req := &api.RequestStopProcessParams{
		Id: cli.Id,
	}
	SendRequest(client, 1, req)
	res, _ := ReadResponse(client)

	if res.Result != nil && res.Result.Success != nil {
		fmt.Printf("Process stopped %s", cli.Id)
	}
}

func DeleteProcess(cli *Delete) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}
	defer client.Close()

	req := &api.RequestDeleteProcessParams{
		Id: cli.Id,
	}
	SendRequest(client, 1, req)
	res, _ := ReadResponse(client)

	if res.Result != nil && res.Result.Success != nil {
		fmt.Printf("Process deleted %s", cli.Id)
	}
}

func RequestStopService() {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}
	defer client.Close()

	req := &api.RequestStopServiceParams{}
	SendRequest(client, 1, req)
	res, _ := ReadResponse(client)

	if res.Result != nil && res.Result.Success != nil {
		fmt.Printf("Requested service shutdown")
	}
}

func SendRequest(client *ServiceConnection, msgID int, req api.RequestParams) error {
	bReq, err := api.NewRequest(1, req)
	if err != nil {
		panic("Could not build request")
	}

	err = client.WriteMsg(bReq)
	if err != nil {
		log.Fatalf("Unknown writing to server: %v", err)
	}

	return err
}

func ReadResponse(client *ServiceConnection) (api.Response, error) {
	data, err := client.ReadMsg()

	if err != nil {
		log.Fatalf("Unknown error reading from server: %v", err)
	}

	res, err := api.UnmarshalResponse(data)

	if err != nil {
		log.Fatalf("Could not decode response: %v", err)
	}

	if res.Error != nil {
		log.Fatalf("Error while doing. %d %s", res.Error.Code, res.Error.Message)
	}

	return res, err
}
