package nodes

import (
	"sync"
	"github.com/butlermatt/dslink"
	"errors"
)

const (
	MaxInt32 = 1<<31 - 1
)

type Requester struct {
	rMu sync.Mutex
	rid int32
	cMu sync.RWMutex
	cache map[int32]chan *dslink.Response
	c   chan<- *dslink.Request
}

func NewRequester(reqChan chan<-*dslink.Request) *Requester {
	return &Requester{c: reqChan, cache: make(map[int32]chan *dslink.Response)}
}

func (r *Requester) HandleResponse(resp *dslink.Response) {
	r.cMu.RLock()
	c := r.cache[resp.Rid]
	r.cMu.RUnlock()

	c<- resp
	if resp.Stream == dslink.StreamClosed {
		r.deleteRid(resp.Rid)
	}
}

func (r *Requester) SendRequest(req *dslink.Request, c chan *dslink.Response) {
	r.cMu.Lock()
	r.cache[req.Rid] = c
	r.cMu.Unlock()

	r.c <- req
	dslink.Log.Println("Sent request")
}

func (r *Requester) CloseRequest(rid int32) {
	req := dslink.NewReq(rid, dslink.MethodClose)
	r.c <- req
	r.deleteRid(rid)
}

func (r *Requester) deleteRid(rid int32) {
	close(r.cache[rid])
	r.cMu.Lock()
	defer r.cMu.Unlock()
	delete(r.cache, rid)
}

func (r *Requester) getRid() int32 {
	r.rMu.Lock()
	defer r.rMu.Unlock()

	if r.rid == int32(MaxInt32) {
		r.rid = 0
	}
	r.rid += 1

	return r.rid
}

func (r *Requester) GetRemoteNode(path string) (dslink.Node, error) {
	// TODO Should have remoteNode type?
	dslink.Log.Println("In GetRemoteNode")
	req := dslink.NewReq(r.getRid(), dslink.MethodList)
	rChan := make(chan *dslink.Response)
	req.Path = path

	r.SendRequest(req, rChan)

	resp := <-rChan

	// TODO Check resp for errors!
	if resp.Error != nil {
		return nil, errors.New("An error was returned")
	}

	nd := NewRemoteNode(path)
	for _, u := range resp.Updates {
		lu, ok := u.([]interface{})
		if !ok {
			dslink.Log.Println("Update isn't a list")
			continue
		}

		n := lu[0].(string)
		dslink.Log.Printf("%s: %v", n, lu[1])
		if n[0] == '$' {
			nd.SetConfig(dslink.NodeConfig(n), lu[1])
		} else if n[0] == '@' {
			nd.SetAttribute(n, lu[1])
		} else {
			// Should be children.
			mp, ok := lu[1].(map[interface{}]interface{})
			if !ok {
				dslink.Log.Printf("Can't convert child %q to node map %#v\n", n, lu[1])
				continue
			}

			c := NewRemoteFromMap(n, mp)
			nd.AddChild(c)
		}
		if n == "$disconnectedTs" {
			return nil, errors.New("No such node")
		}
		if n == "$is" {
			isT, _ := lu[1].(string)
			if isT == "node" {
				continue
			}

			req := dslink.NewReq(r.getRid(), dslink.MethodList)
			req.Path = "/def/profile/" + isT
			isChan := make(chan *dslink.Response)

			r.SendRequest(req, isChan)
			resp := <-isChan
			r.CloseRequest(req.Rid)
			for _, u := range resp.Updates {
				up, ok := u.([]interface{})
				if !ok {
					dslink.Log.Printf("Unable to convert %#v to slice", u)
					continue
				}
				n, _ := up[0].(string)
				if n == "$is" {
					if up[1] == "node" {
						continue
					} else {
						dslink.Log.Printf("$is on Profile is not node: %q\n", up[1])
					}
				} else if n[0] == '$' {
					nd.SetConfig(dslink.NodeConfig(n), up[1])
				} else if n[0] == '@' {
					nd.SetAttribute(n, up[1])
				}
			}
		}
	}

	r.CloseRequest(req.Rid)

	return nd, nil
}