package client

import (
	"errors"
	"jstarpl/jpm/api"
	"time"

	ipc "github.com/james-barrow/golang-ipc"
)

type ServiceConnection struct {
	client *ipc.Client
}

const serviceConnectionTimeout = 3 * time.Second

var (
	ErrServiceConnection        = errors.New("could not connect to service")
	ErrServiceConnectionTimeout = errors.New("could not connect to service, waiting for Connected status timed out")
	ErrInvalidMessage           = errors.New("could not read from service connection")
)

func DialService() (*ServiceConnection, error) {
	client, err := ipc.StartClient(api.IPCName, nil)
	if err != nil {
		return nil, ErrServiceConnection
	}

	start := time.Now()

	for {
		if time.Since(start) > serviceConnectionTimeout {
			return nil, ErrServiceConnectionTimeout
		}

		data, err := client.Read()
		if err != nil {
			return nil, ErrServiceConnection
		}

		if data.Status == "Connected" {
			break
		}
	}

	return &ServiceConnection{
		client: client,
	}, nil
}

func (c *ServiceConnection) WriteMsg(message []byte) error {
	return c.client.Write(api.MsgType, message)
}

func (c *ServiceConnection) ReadMsg() ([]byte, error) {
	for {
		msg, err := c.client.Read()
		if err != nil {
			return msg.Data, ErrInvalidMessage
		}
		if msg.MsgType == api.MsgType {
			return msg.Data, nil
		}
	}
}

func (c *ServiceConnection) Close() {
	c.client.Close()
}
