package client

import (
	"jstarpl/jpm/api"
	"log"

	ipc "github.com/james-barrow/golang-ipc"
)

func ListProcesses() {
	client, err := ipc.StartClient("jpm-ipc", nil)
	if err != nil {
		panic("Could not open `jpm-ipc` IPC channel. Check if the service is running.")
	}

	for {
		data, err := client.Read()
		if err != nil {
			log.Fatalf("Unknown connection error: %v", err)
		}

		if data.Status == "Connected" {
			break
		}
	}

	req := &api.RequestListProcessesParams{}
	bReq, err := api.NewRequest(1, req)
	if err != nil {
		panic("Could not build request")
	}

	err = client.Write(1, bReq)
	if err != nil {
		log.Fatalf("Unknown writing to server: %v", err)
	}

	data, err := client.Read()

	if err != nil {
		log.Fatalf("Unknown error reading from server: %v", err)
	}

	log.Default().Printf("Message received: %d Length %d %v", data.MsgType, len(data.Data), string(data.Data[:]))

	client.Close()
}
