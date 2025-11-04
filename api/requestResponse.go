package api

import (
	"encoding/json"
	"errors"
)

type MethodName string

const MsgType int = 100

const IPCName string = "jpm-ipc"

const (
	ListProcesses      MethodName = "listProcesses"
	StartProcess       MethodName = "startProcess"
	StopProcess        MethodName = "stopProcess"
	DeleteProcess      MethodName = "deleteProcess"
	RequestStopService MethodName = "requestStopService"
)

type JSONRPCErrors int

const (
	ParseError     JSONRPCErrors = -32700
	InvalidRequest JSONRPCErrors = -32600
	MethodNotFound JSONRPCErrors = -32601
	InvalidParams  JSONRPCErrors = -32602
	InternalError  JSONRPCErrors = -32603
	ServerError    JSONRPCErrors = -32000
)

type RequestListProcessesParams struct{}

func (r RequestListProcessesParams) Type() MethodName {
	return ListProcesses
}

type RequestStartProcessParams struct {
	Name string   `json:"name,omitempty"`
	Exec string   `json:"exec"`
	Arg  []string `json:"args"`
	Env  []string `json:"env"`
	Dir  string   `json:"cwd"`
}

func (r RequestStartProcessParams) Type() MethodName {
	return StartProcess
}

type RequestStopProcessParams struct {
	Id    string `json:"id,omitempty"`
	Query string `json:"query,omitempty"`
}

func (r RequestStopProcessParams) Type() MethodName {
	return StopProcess
}

type RequestDeleteProcessParams struct {
	Id    string `json:"id,omitempty"`
	Query string `json:"query,omitempty"`
}

func (r RequestDeleteProcessParams) Type() MethodName {
	return DeleteProcess
}

type RequestStopServiceParams struct {
}

func (r RequestStopServiceParams) Type() MethodName {
	return RequestStopService
}

type RequestParams interface {
	Type() MethodName
}

type Request struct {
	Header string          `json:"jsonrpc"`
	Method MethodName      `json:"method"`
	MsgID  int             `json:"id"`
	Params json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	Header string          `json:"jsonrpc"`
	Result *ResponseResult `json:"result,omitempty"`
	Error  *ResponseError  `json:"params,omitempty"`
	MsgID  int             `json:"id"`
}

type Process struct {
	Id       string   `json:"id"`
	Name     string   `json:"name,omitempty"`
	Exec     string   `json:"exec"`
	Arg      []string `json:"args"`
	Env      []string `json:"env"`
	Dir      string   `json:"cwd"`
	Uptime   int      `json:"uptime"`
	Status   Status   `json:"status"`
	ExitCode int      `json:"exitCode"`
}

type ResponseResult struct {
	Success     *string      `json:"success,omitempty"`
	ProcessList *([]Process) `json:"processList,omitempty"`
	ProcessId   *string      `json:"processId,omitempty"`
}

type ResponseError struct {
	Code    int              `json:"code"`
	Message string           `json:"string"`
	Data    *json.RawMessage `json:"data,omitempty"`
}

func NewRequest(msgID int, params RequestParams) ([]byte, error) {
	if params == nil {
		return nil, errors.New("message must not be nil")
	}

	messageType := params.Type()
	dataJson, _ := json.Marshal(params)

	envelope := Request{Header: "2.0", Method: messageType, MsgID: msgID, Params: dataJson}
	return json.Marshal(envelope)
}

func NewSuccessResponse(msgID int, data *ResponseResult) ([]byte, error) {
	envelope := Response{Header: "2.0", MsgID: msgID, Result: data}
	return json.Marshal(envelope)
}

func NewErrorResponse(msgID int, code int, message string) ([]byte, error) {
	envelope := Response{Header: "2.0", MsgID: msgID, Error: &ResponseError{Code: code, Message: message}}
	return json.Marshal(envelope)
}

func UnmarshalResponse(data []byte) (Response, error) {
	res := Response{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
