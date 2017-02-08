package nodes

import (
	"github.com/butlermatt/dslink"
)

type SimpleProvider struct {
	cache    map[string]*SimpleNode
	root     *SimpleNode
	resp     chan<- *dslink.Response
	listResp map[int32]*SimpleNode
}

func (s *SimpleProvider) GetNode(path string) (dslink.Node, bool) {
	nd, ok := s.cache[path]
	return nd, ok
}

func (s *SimpleProvider) GetRoot() dslink.Node {
	return s.root
}

func NewProvider(resp chan<- *dslink.Response) *SimpleProvider {
	sp := &SimpleProvider{cache: make(map[string]*SimpleNode), listResp: make(map[int32]*SimpleNode)}
	r := NewNode("", sp)
	sp.root = r
	sp.cache["/"] = r
	sp.resp = resp
	return sp
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
	nd.Close(req)
	delete(s.listResp, req.Rid)
}