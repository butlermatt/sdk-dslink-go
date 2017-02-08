package nodes

import (
	"github.com/butlermatt/dslink"
)

type SimpleProvider struct {
	cache    map[string]dslink.Node
	root     dslink.Node
	resp     chan<- *dslink.Response
	listResp map[int32]dslink.Node
}

func (s *SimpleProvider) GetNode(path string) (dslink.Node, bool) {
	nd, ok := s.cache[path]
	return nd, ok
}

func (s *SimpleProvider) GetRoot() dslink.Node {
	return s.root
}

func NewProvider(resp chan<- *dslink.Response) *SimpleProvider {
	sp := &SimpleProvider{cache: make(map[string]dslink.Node), listResp: make(map[int32]dslink.Node)}
	r := NewNode("", sp)
	sp.root = r
	sp.cache["/"] = r
	sp.resp = resp
	return sp
}

func (s *SimpleProvider) AddNode(path string, node dslink.Node) {
	s.cache[path] = node
}

func (s *SimpleProvider) RemoveNode(path string) dslink.Node {
	nd := s.cache[path]
	if nd != nil {
		nd.Remove()
	}

	return nd
}

func (s *SimpleProvider) SendResponse(resp *dslink.Response) {
	s.resp<- resp
}

func (s *SimpleProvider) HandleRequest(req *dslink.Request) *dslink.Response {
	dslink.Log.Printf("Received Request: %+v", req)

	switch req.Method {
	case "list":
		return s.handleList(req)
	case "close":
		s.handleClose(req)
	}
	return nil
}

func (s *SimpleProvider) handleList(req *dslink.Request) *dslink.Response {
	nd := s.cache[req.Path]
	s.listResp[req.Rid] = nd

	return nd.List(req)
}

func (s *SimpleProvider) handleClose(req *dslink.Request) {
	nd := s.listResp[req.Rid]
	if nd != nil {
		nd.Close(req)
	}
	delete(s.listResp, req.Rid)
}