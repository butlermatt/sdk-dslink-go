package nodes

import (
	"errors"
	"sync"
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/log"
)

type LocalNode struct {
	provider    *Provider
	attr        map[string]interface{}
	conf        map[dslink.NodeConfig]interface{}
	Parent      *LocalNode
	name        string
	path        string
	vMu         sync.RWMutex
	value       interface{}
	valType     dslink.ValueType
	onInvoke    dslink.InvokeFn
	columns     []map[string]interface{}
	chld        map[string]*LocalNode
	sMu         sync.RWMutex
	subscribers []int32
	lMu         sync.RWMutex
	listSubs    []int32
	onSet       dslink.OnSetValue
}

func (n *LocalNode) Name() string {
	return n.name
}

func (n *LocalNode) Attributes() map[string]interface{} {
	return n.attr
}

func (n *LocalNode) GetAttribute(name string) (interface{}, bool) {
	a, ok := n.attr[name]
	return a, ok
}

func (n *LocalNode) SetAttribute(name string, v interface{}) {
	n.attr[name] = v
}

func (n *LocalNode) Configs() map[dslink.NodeConfig]interface{} {
	return n.conf
}

func (n *LocalNode) GetConfig(name dslink.NodeConfig) (interface{}, bool) {
	c, ok := n.conf[name]
	return c, ok
}

func (n *LocalNode) SetConfig(name dslink.NodeConfig, value interface{}) {
	n.conf[name] = value
}

func (n *LocalNode) Children() map[string]*LocalNode {
	return n.chld
}

func (n *LocalNode) GetChild(name string) dslink.Node {
	return n.chld[name]
}

func (n *LocalNode) AddChild(node dslink.Node) error {
	sn, ok := node.(*LocalNode)
	if !ok {
		return errors.New("Can't add unknown node type")
	}
	sn.Parent = n
	sn.path = n.path + "/" + sn.name
	n.provider.AddNode(sn.path, sn)
	n.chld[sn.name] = sn

	n.notifyList(sn.name, sn.ToMap())

	return nil
}

func (n *LocalNode) Remove() {
	p := n.Parent
	n.Parent = nil

	if p != nil {
		p.RemoveChild(n.name)
	}

	for name, c := range n.chld {
		c.Remove()
		delete(n.chld, name)
	}

	prov := n.provider
	n.provider = nil
	if prov != nil {
		prov.RemoveNode(n.path)
	}
}

func (n *LocalNode) RemoveChild(name string) dslink.Node {
	nd := n.chld[name]
	delete(n.chld, name)

	if nd != nil {
		nd.Remove()
		n.lMu.RLock()
		defer n.lMu.RUnlock()
		for _, i := range n.listSubs {
			r := dslink.NewResp(i)
			r.Updates = append(r.Updates, map[string]string{"name": name, "change": "remove"})
			n.provider.SendResponse(r)
		}
	}

	return nd
}

func (n *LocalNode) notifyList(name string, value interface{}) {
	n.lMu.RLock()
	defer n.lMu.RUnlock()
	for _, i := range n.listSubs {
		r := &dslink.Response{Rid: i}
		r.AddUpdate(name, value)
		n.provider.SendResponse(r)
	}
}

func (n *LocalNode) notifySubs(update *dslink.ValueUpdate) {
	n.sMu.RLock()
	defer n.sMu.RUnlock()
	if len(n.subscribers) <= 0 {
		return
	}

	r := dslink.NewResp(0)

	for _, i := range n.subscribers {
		r.AddUpdate(i, update)
	}
	n.provider.SendResponse(r)
}

func (n *LocalNode) List(request *dslink.Request) *dslink.Response {
	n.lMu.Lock()
	n.listSubs = append(n.listSubs, request.Rid)
	n.lMu.Unlock()

	r := dslink.NewResp(request.Rid)
	r.Stream = dslink.StreamOpen

	is, _ := n.GetConfig(dslink.ConfigIs)
	r.AddUpdate(dslink.ConfigIs, is)

	for k, v := range n.conf {
		if k == dslink.ConfigIs {
			continue
		}
		r.AddUpdate(k, v)
	}

	for k, v := range n.attr {
		r.AddUpdate(k, v)
	}

	for name, nd := range n.chld {
		r.AddUpdate(name, nd.ToMap())
	}

	return r
}

func (n *LocalNode) Close(request *dslink.Request) {
	i := -1
	n.lMu.Lock()
	defer n.lMu.Unlock()
	for j, id := range n.listSubs {
		if id == request.Rid {
			i = j
			break
		}
	}

	if i != -1 {
		n.listSubs[i] = n.listSubs[len(n.listSubs) - 1]
		n.listSubs = n.listSubs[:len(n.listSubs) - 1]
		log.Debug.Printf("Closed conn for Rid: %d\n", request.Rid)
	}
}

func (n *LocalNode) Subscribe(sid int32) {
	n.sMu.Lock()
	defer n.sMu.Unlock()
	n.subscribers = append(n.subscribers, sid)
}

func (n *LocalNode) Unsubscribe(sid int32) {
	i := -1

	n.sMu.Lock()
	defer n.sMu.Unlock()
	for j, id := range n.subscribers {
		if id == sid {
			i = j
			break
		}
	}

	if i != -1 {
		n.subscribers[i] = n.subscribers[len(n.subscribers) - 1]
		n.subscribers = n.subscribers[:len(n.subscribers) - 1]
	}
}

func (n *LocalNode) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m[string(dslink.ConfigIs)] = n.conf[dslink.ConfigIs]
	name, ok := n.conf[dslink.ConfigName]
	if ok {
		m[string(dslink.ConfigName)] = name
	}
	perm, ok := n.conf[dslink.ConfigPermission]
	if ok && perm != nil && perm != dslink.PermRead {
		m[string(dslink.ConfigPermission)] = perm
	}
	if n.valType != "" {
		m[string(dslink.ConfigType)] = n.valType
	}
	if n.conf[dslink.ConfigInterface] != nil {
		m[string(dslink.ConfigInterface)] = n.conf[dslink.ConfigInterface]
	}
	if n.conf[dslink.ConfigInvokable] != nil {
		m[string(dslink.ConfigInvokable)] = n.conf[dslink.ConfigInvokable]
	}

	return m
}

func (n *LocalNode) GetType() dslink.ValueType {
	return n.valType
}

func (n *LocalNode) SetType(t dslink.ValueType) {
	n.conf[dslink.ConfigType] = t
	n.valType = t
}

func (n *LocalNode) AddAction(fn dslink.InvokeFn, params []dslink.Params, cols []dslink.Column, result string) {
	n.onInvoke = fn
	var p []map[string]interface{}
	for _, v := range params {
		m := make(map[string]interface{})
		for k, j := range v {
			m[string(k)] = j
		}
		p = append(p, m)
	}
	n.conf[dslink.ConfigParams] = p

	var columns []map[string]interface{}
	for _, c := range cols {
		m := make(map[string]interface{})
		m[string(dslink.ParamName)] = c.Name
		m[string(dslink.ParamType)] = string(c.Type)
		if c.Default != nil {
			m[string(dslink.ParamDef)] = c.Default
		}
		columns = append(columns, m)
	}
	n.columns = columns
	n.conf[dslink.ConfigColumns] = columns
	n.conf[dslink.ConfigInvokable] = dslink.PermWrite
	n.conf[dslink.ConfigResult] = result
}

func (n *LocalNode) UpdateValue(v interface{}) {
	n.vMu.Lock()
	n.value = v
	n.vMu.Unlock()
	// TODO: Something about the subscription and stuff
	val := dslink.NewValueUpdate(v)
	n.notifySubs(val)
}

func (n *LocalNode) Value() interface{} {
	n.vMu.RLock()
	defer n.vMu.RUnlock()
	return n.value
}

func (n *LocalNode) Invoke(req *dslink.Request) {
	r := dslink.NewResp(req.Rid)

	perm := dslink.PermType(req.Permit)
	if perm == "" || perm.Level() == -1 {
		perm = dslink.PermConfig
	}

	pr, ok := n.GetConfig(dslink.ConfigInvokable)
	prs, _ := pr.(string)
	if !ok {
		r.Error = dslink.ErrInvalidMethod
		n.provider.SendResponse(r)
		return
	}

	permReq := dslink.PermType(prs)
	if perm.Level() < permReq.Level() {
		r.Error = dslink.ErrPermissionDenied
		n.provider.SendResponse(r)
		return
	}

	r.Columns = n.columns

	if n.onInvoke == nil {
		empty := []interface{}{}
		r.Updates = append(r.Updates, empty)
		n.provider.SendResponse(r)
		return
	}
	rType, _ := n.GetConfig(dslink.ConfigResult)
	s, _ := rType.(string)

	retChan := make(chan []interface{})
	go n.onInvoke(req.Params, retChan)

	if s != dslink.ResultStream {
		r.Stream = dslink.StreamClosed
		for u := range retChan {
			r.Updates = append(r.Updates, u)
		}
		n.provider.SendResponse(r)
		return
	}

	var up [][]interface{}
	r.Stream = dslink.StreamOpen
	for {
		select {
		case data, ok := <-retChan:
			if !ok {
				retChan = nil
			} else {
				up = append(up, data)
			}
		default:
			if len(up) == 0 {
				continue
			}
			for _, u := range up {
				r.Updates = append(r.Updates, u)
			}
			up = up[:0]
			n.provider.SendResponse(r)
			r = dslink.NewResp(req.Rid)
			//r.Stream = dslink.StreamOpen

		}
		if retChan == nil {
			break
		}
	}

	if len(up) > 0 {
		r.Stream = dslink.StreamClosed
		for _, u := range up {
			r.Updates = append(r.Updates, u)
		}
		n.provider.SendResponse(r)
	}
}

func (n *LocalNode) Set(req *dslink.Request) *dslink.MsgErr {
	perm := dslink.PermType(req.Permit)
	if perm == "" || perm.Level() == -1 {
		perm = dslink.PermConfig
	}

	pr, ok := n.GetConfig(dslink.ConfigWritable)
	prs, _ := pr.(string)
	if !ok {
		return dslink.ErrInvalidValue
	}

	permReq := dslink.PermType(prs)
	if perm.Level() < permReq.Level() {
		return dslink.ErrPermissionDenied
	}

	ok = n.onSet(n, req.Value)
	if !ok {
		return nil
	}

	n.UpdateValue(req.Value)

	return nil
}

func (n *LocalNode) EnableSet(perm dslink.PermType, onSet dslink.OnSetValue) {
	n.conf[dslink.ConfigWritable] = perm
	n.onSet = onSet
}

func NewNode(name string, provider *Provider) *LocalNode {
	sn := &LocalNode{
		name:     name,
		provider: provider,
		attr:     make(map[string]interface{}),
		conf:     make(map[dslink.NodeConfig]interface{}),
		chld:     make(map[string]*LocalNode),
		sMu:      sync.RWMutex{},
		lMu:      sync.RWMutex{},
	}

	sn.conf[dslink.ConfigIs] = "node"

	return sn
}
