package nodes

import (
	"sync"
	"github.com/butlermatt/dslink"
)

type Requester struct {
	rMu sync.Mutex
	rid int32
	c   chan<- *dslink.Request
}

func NewRequester(reqChan chan<-*dslink.Request) *Requester {
	return &Requester{c: reqChan}
}

const (
	MaxInt32 = 1<<31 - 1
)

func (r *Requester) getRid() int32 {
	r.rMu.Lock()
	defer r.rMu.Unlock()
	if r.rid == MaxInt32 {
		r.rid = 0
	}
	r.rid += 1

	return r.rid
}

func (r *Requester) GetRemoteNode(path string) dslink.Node {
	// TODO Should have remoteNode type?
	req := dslink.NewReq(r.getRid(), dslink.MethodList)
	req.Path = path

	return nil
}