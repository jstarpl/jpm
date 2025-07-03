package api

import (
	"encoding/json"
	"errors"
)

type MethodName string

const (
	ListProcesses MethodName = "listProcesses"
	StartProcess  MethodName = "startProcess"
	StopProcess   MethodName = "stopProcess"
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
	Id int
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

type ResponseParams interface {
}

type Response struct {
	Header string          `json:"jsonrpc"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ResponseError  `json:"params,omitempty"`
	MsgID  int             `json:"id"`
}

type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"string"`
	Data    json.RawMessage `json:"data,omitempty"`
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

func NewSuccessResponse(msgID int, data ResponseParams) ([]byte, error) {
	dataJson, _ := json.Marshal(data)

	envelope := Response{Header: "2.0", MsgID: msgID, Result: dataJson}
	return json.Marshal(envelope)
}

func NewErrorResponse(msgID int, code int, message string) ([]byte, error) {
	envelope := Response{Header: "2.0", MsgID: msgID, Error: &ResponseError{Code: code, Message: message}}
	return json.Marshal(envelope)
}
