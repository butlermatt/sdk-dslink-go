package nodes

import (
	"sync"
	"github.com/butlermatt/dslink"
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
	dslink.Log.Printf("Returned response is: %v", resp)

	for _, u := range resp.Updates {
		lu, ok := u.([]interface{})
		if !ok {
			dslink.Log.Println("Update isn't a list")
			continue
		}

		n := lu[0].(string)
		dslink.Log.Printf("%s: %v", n, lu[1])
	}

	// TODO Need to send a close as well
	r.CloseRequest(req.Rid)

	return nil, nil
}