package dslink

//
//type Provider interface {
//	// TODO
//	GetNode(path string) Node
//	GetRoot() Node
//	AddNode(path string, node Node)
//	RemoveNode(path string) Node
//}

type Responder interface {
	HandleRequest(req *Request) *Response
	SendResponse(resp *Response)
}

type Requester interface {
	HandleResponse(*Response)
	SendRequest(*Request, chan *Response)
}