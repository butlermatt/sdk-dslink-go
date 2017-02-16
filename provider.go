package dslink

import (
	lg "log"
)

var Log *lg.Logger

type Provider interface {
	// TODO
	GetNode(path string) Node
	GetRoot() Node
	HandleRequest(req *Request) *Response
	SendResponse(resp *Response)
	AddNode(path string, node Node)
	RemoveNode(path string) Node
	//GetOrCreateNode(path string) Node;
}

type Requester interface {
	// TODO
	HandleResponse(*Response)
	SendRequest(*Request, chan *Response)
	CloseRequest(int32)
	GetRemoteNode(string) (Node, error)
}