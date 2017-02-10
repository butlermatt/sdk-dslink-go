package nodes

import (
	"sync"
	"github.com/butlermatt/dslink"
)

type SimpleProvider struct {
	root        dslink.Node
	resp        chan<- *dslink.Response
	cMu         sync.Mutex
	cache       map[string]dslink.Node
	lMu         sync.Mutex
	listResp    map[int32]dslink.Node
	sMu         sync.Mutex
	subscribers map[int32]dslink.Valued
}

func (s *SimpleProvider) GetNode(path string) (dslink.Node, bool) {
	s.cMu.Lock()
	defer s.cMu.Unlock()
	nd, ok := s.cache[path]
	return nd, ok
}

func (s *SimpleProvider) GetRoot() dslink.Node {
	return s.root
}

func (s *SimpleProvider) AddNode(path string, node dslink.Node) {
	s.cMu.Lock()
	defer s.cMu.Unlock()
	s.cache[path] = node
}

func (s *SimpleProvider) RemoveNode(path string) dslink.Node {
	s.cMu.Lock()
	nd := s.cache[path]
	s.cMu.Unlock()
	if nd != nil {
		nd.Remove()
	}

	return nd
}

func (s *SimpleProvider) SendResponse(resp *dslink.Response) {
	s.resp <- resp
}

func (s *SimpleProvider) HandleRequest(req *dslink.Request) *dslink.Response {
	dslink.Log.Printf("Received Request: %+v", req)

	switch req.Method {
	case dslink.MethodList:
		return s.handleList(req)
	case dslink.MethodClose:
		s.handleClose(req)
	case dslink.MethodSub:
		return s.handleSub(req)
	case dslink.MethodUnsub:
		return s.handleUnsub(req)
	case dslink.MethodInvoke:
		s.handleInvoke(req)
	default:
		dslink.Log.Printf("Unhandled method: %s", req.Method)
	}
	return nil
}

func (s *SimpleProvider) handleList(req *dslink.Request) *dslink.Response {
	s.cMu.Lock()
	nd := s.cache[req.Path]
	s.cMu.Unlock()

	s.lMu.Lock()
	s.listResp[req.Rid] = nd
	s.lMu.Unlock()

	if nd == nil {
		r := dslink.NewResp(req.Rid)
		r.AddUpdate("$is", "node")
		return r
	}

	return nd.List(req)
}

func (s *SimpleProvider) handleClose(req *dslink.Request) {
	s.lMu.Lock()
	defer s.lMu.Unlock()
	nd := s.listResp[req.Rid]
	if nd != nil {
		nd.Close(req)
	}
	delete(s.listResp, req.Rid)
}

func (s *SimpleProvider) handleSub(req *dslink.Request) *dslink.Response {
	r := dslink.NewResp(req.Rid)
	r.Stream = dslink.StreamClosed

	var newSubs []int32
	for _, p := range req.Paths {
		s.cMu.Lock()
		n := s.cache[p.Path]
		s.cMu.Unlock()

		v, ok := n.(dslink.Valued)
		if ok {
			newSubs = append(newSubs, p.Sid)
			s.sMu.Lock()
			s.subscribers[p.Sid] = v
			s.sMu.Unlock()
			v.Subscribe(p.Sid)
		} else {
			dslink.Log.Printf("Can't subscribe to \"%s\". Not a value", p.Path)
		}
	}

	r2 := dslink.NewResp(0)
	for _, sid := range newSubs {
		s.sMu.Lock()
		v := s.subscribers[sid]
		s.sMu.Unlock()

		if v != nil {
			vu := dslink.NewValueUpdate(v.Value())
			r2.AddUpdate(sid, vu)
		}
	}

	go s.SendResponse(r2)

	return r
}

func (s *SimpleProvider) handleUnsub(req *dslink.Request) *dslink.Response {
	r := dslink.NewResp(req.Rid)
	r.Stream = dslink.StreamClosed

	for _, i := range req.Sids {
		s.sMu.Lock()
		nd := s.subscribers[i]
		if nd != nil {
			nd.Unsubscribe(i)
		}
		delete(s.subscribers, i)
		s.sMu.Unlock()
	}

	return r
}

func (s *SimpleProvider) handleInvoke(req *dslink.Request) {
	s.cMu.Lock()
	n := s.cache[req.Path]
	s.cMu.Unlock()
	in, ok := n.(dslink.Invokable)
	if ok {
		go in.Invoke(req)
	}
}

func NewProvider(resp chan<- *dslink.Response) *SimpleProvider {
	sp := &SimpleProvider{
		cache:       make(map[string]dslink.Node),
		listResp:    make(map[int32]dslink.Node),
		subscribers: make(map[int32]dslink.Valued),
		sMu:         sync.Mutex{},
		lMu:         sync.Mutex{},
		cMu:         sync.Mutex{},
	}
	r := NewNode("", sp)
	sp.root = r
	sp.cache["/"] = r
	sp.resp = resp
	return sp
}

