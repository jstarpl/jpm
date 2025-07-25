package client

import (
	"jstarpl/jpm/api"
	"log"
	"os"
)

type Ps struct{}

type Start struct {
	Name string   `name:"name" help:"Name of the process"`
	Args []string `arg:""`
}

type Stop struct {
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

	if res.Result != nil && res.Result.ProcessList != nil {
		log.Default().Printf("Process list %v", string(data))
	}

	client.Close()
}

func StartProcess(cli *Start) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}

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

	if res.Result != nil && res.Result.Success != nil {
		log.Default().Printf("Process started %v", string(data))
	}

	client.Close()
}

func StopProcess(cli *Stop) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}

	req := &api.RequestStopProcessParams{
		Id: cli.Id,
	}
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

	if res.Result != nil && res.Result.Success != nil {
		log.Default().Printf("Process stopped %v", string(data))
	}

	client.Close()
}

func DeleteProcess(cli *Delete) {
	client, err := DialService()
	if err != nil {
		log.Fatalf("Could not connect to service: %v", err)
	}

	req := &api.RequestDeleteProcessParams{
		Id: cli.Id,
	}
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

	if res.Result != nil && res.Result.Success != nil {
		log.Default().Printf("Process deleted %v", string(data))
	}

	client.Close()
}
