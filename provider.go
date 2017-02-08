package dslink

import (
	lg "log"
)

var Log *lg.Logger

type Provider interface {
	// TODO
	GetNode(path string) (Node, bool)
	GetRoot() Node
	HandleRequest(req *Request) *Response
	SendResponse(resp *Response)
	AddNode(path string, node Node)
	RemoveNode(path string) Node
	//GetOrCreateNode(path string) Node;
}
