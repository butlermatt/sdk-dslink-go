package nodes

import (
	"sync"
	"github.com/butlermatt/dslink"
)

type SimpleProvider struct {
	root        dslink.Node
	c           chan<- *dslink.Response
	lMu         sync.Mutex
	listResp    map[int32]dslink.Node
	cMu         sync.RWMutex
	cache       map[string]dslink.Node
	sMu         sync.RWMutex
	subscribers map[int32]dslink.Valued
}

// GetNode will attempt to return the Node located at the Specified path.
func (s *SimpleProvider) GetNode(path string) dslink.Node {
	s.cMu.RLock()
	defer s.cMu.RUnlock()
	nd := s.cache[path]
	return nd
}

// GetRoot returns the root node of this DSLink when run as a Responder.
func (s *SimpleProvider) GetRoot() dslink.Node {
	return s.root
}

// AddNode will add the specified node on the specified path. However it will not establish the appropriate
// parent/child relationship and nodes should be added directly from other nodes.
func (s *SimpleProvider) AddNode(path string, node dslink.Node) {
	s.cMu.Lock()
	defer s.cMu.Unlock()
	s.cache[path] = node
}

// RemoveNode will remove the node at the specified path. It will return the node which was removed. It will
// also attempt to call Remove on the Node itself to ensure the parent/child associations are cleaned up as
// well.
func (s *SimpleProvider) RemoveNode(path string) dslink.Node {
	s.cMu.Lock()
	nd := s.cache[path]
	delete(s.cache, path)
	s.cMu.Unlock()

	if nd != nil {
		nd.Remove()
	}

	return nd
}

// SendResponse is used by provider and node implementations for Responders to send an async response back to the
// remote requester.
func (s *SimpleProvider) SendResponse(resp *dslink.Response) {
	s.c <- resp
}

// HandleRequest must be implemented by a Responder to handle incoming requests. It may return a Response
// directly or it may return nil and send an async response with SendResponse.
func (s *SimpleProvider) HandleRequest(req *dslink.Request) *dslink.Response {
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
	case dslink.MethodSet:
		s.handleSet(req)
	default:
		dslink.Log.Printf("Unhandled method: %s", req.Method)
	}
	return nil
}

func (s *SimpleProvider) handleList(req *dslink.Request) *dslink.Response {
	s.cMu.RLock()
	nd := s.cache[req.Path]
	s.cMu.RUnlock()

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
		s.cMu.RLock()
		n := s.cache[p.Path]
		s.cMu.RUnlock()

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
		s.sMu.RLock()
		v := s.subscribers[sid]
		s.sMu.RUnlock()

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
	s.cMu.RLock()
	n := s.cache[req.Path]
	s.cMu.RUnlock()

	in, ok := n.(dslink.Invokable)
	if !ok {
		r := dslink.NewResp(req.Rid)
		r.Error = dslink.ErrInvalidMethod
		s.c<- r
		return
	}

	go in.Invoke(req)
}

func (s *SimpleProvider) handleSet(req *dslink.Request) {
	s.cMu.RLock()
	n := s.cache[req.Path]
	s.cMu.RUnlock()

	se, ok := n.(dslink.Settable)
	if !ok {
		r := dslink.NewResp(req.Rid)
		r.Error = dslink.ErrInvalidValue
		s.SendResponse(r)
		return
	}

	err := se.Set(req)
	if err != nil {
		r := dslink.NewResp(req.Rid)
		r.Error = err
		s.SendResponse(r)
	}
}

// NewProvider returns a new SimpleProvider which is a simple implementation of the Provider and Node interfaces.
// It receives a Response sending channel to return asynchronous Responses to requests.
func NewProvider(resp chan<- *dslink.Response) *SimpleProvider {
	sp := &SimpleProvider{
		cache:       make(map[string]dslink.Node),
		listResp:    make(map[int32]dslink.Node),
		subscribers: make(map[int32]dslink.Valued),
		lMu:         sync.Mutex{},
		sMu:         sync.RWMutex{},
		cMu:         sync.RWMutex{},
	}
	r := NewNode("", sp)
	sp.root = r
	sp.cache["/"] = r
	sp.c = resp
	return sp
}

