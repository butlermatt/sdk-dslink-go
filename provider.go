package dslink

import (
	lg "log"
)

var Log *lg.Logger

type Provider interface {
	// TODO
	GetNode(path string) (Node, bool)
	GetRoot() Node
	HandleRequest(req Request)
	//GetOrCreateNode(path string) Node;
}
