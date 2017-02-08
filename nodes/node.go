package nodes

import (
	"errors"
	"github.com/butlermatt/dslink"
)

type SimpleNode struct {
	p	    dslink.Provider
	attr        map[string]interface{}
	conf        map[string]interface{}
	chld        map[string]dslink.Node
	Parent      dslink.Node
	name        string
	displayName string
	path        string
	listSubs    []int32
	subscribers []int32
	value       interface{}
	valType     dslink.ValueType
}

func (n *SimpleNode) GetAttribute(name string) (interface{}, bool) {
	a, ok := n.attr[name]
	return a, ok
}

func (n *SimpleNode) GetConfig(name string) (interface{}, bool) {
	c, ok := n.conf[name]
	return c, ok
}

func (n *SimpleNode) GetChild(name string) dslink.Node {
	return n.chld[name]
}

func (n *SimpleNode) AddChild(node dslink.Node) error {
	sn, ok := node.(*SimpleNode)
	if !ok {
		return errors.New("Can't add unknown node type")
	}
	sn.Parent = n
	sn.path = n.path + "/" + sn.name
	n.p.AddNode(sn.path, sn)
	n.chld[sn.name] = sn

	n.notifyList(sn.name, sn.ToMap())

	return nil
}

func (n *SimpleNode) Remove() {
	p := n.Parent
	n.Parent = nil

	if p != nil {
		p.RemoveChild(n.name)
	}

	for name, c := range n.chld {
		c.Remove()
		delete(n.chld, name)
	}

	prov := n.p
	n.p = nil
	if prov != nil {
		prov.RemoveNode(n.path)
	}
}

func (n *SimpleNode) RemoveChild(name string) dslink.Node {
	nd := n.chld[name]
	delete(n.chld, name)

	if nd != nil {
		nd.Remove()
	}

	return nd
}

func (n *SimpleNode) notifyList(name string, value interface{}) {
	dslink.Log.Printf("There are %d nodes to notify\n", len(n.listSubs))
	for _, i := range n.listSubs {
		r := &dslink.Response{Rid: i}
		r.AddUpdate(name, value)
		n.p.SendResponse(r)
	}
}

func (n *SimpleNode) notifySubs(update *dslink.ValueUpdate) {
	dslink.Log.Printf("There are %d nodes to notify\n", len(n.subscribers))
	if len(n.subscribers) <= 0 {
		return
	}

	r := dslink.NewResp(0)
	for _, i := range n.subscribers {
		r.AddUpdate(i, update)
	}
	n.p.SendResponse(r)
}

func (n *SimpleNode) List(request *dslink.Request) *dslink.Response {
	n.listSubs = append(n.listSubs, request.Rid)
	r := dslink.NewResp(request.Rid)
	r.Stream = "open"

	is, _ := n.GetConfig(`$is`)
	r.AddUpdate(`$is`, is)


	for name, nd := range n.chld {
		r.AddUpdate(name, nd.ToMap())
	}

	return r
}

func (n *SimpleNode) Close(request *dslink.Request) {
	i := -1
	for j, id := range n.listSubs {
		if id == request.Rid {
			i = j
			break
		}
	}

	if i != -1 {
		n.listSubs[i] = n.listSubs[len(n.listSubs) - 1]
		n.listSubs = n.listSubs[:len(n.listSubs) - 1]
		dslink.Log.Printf("Closed link for Rid: %d\n", request.Rid)
	}
}

func (n *SimpleNode) Subscribe(sid int32) {
	n.subscribers = append(n.subscribers, sid)
}

func (n *SimpleNode) Unsubscribe(sid int32) {
	i := -1
	for j, id := range n.subscribers {
		if id == sid {
			i = j
			break
		}
	}

	if i != -1 {
		n.subscribers[i] = n.subscribers[len(n.subscribers) - 1]
		n.subscribers = n.subscribers[:len(n.subscribers) - 1]
		dslink.Log.Printf("Closed stream for Sid: %d\n", sid)
	}
}

func (n *SimpleNode) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m[`$is`] = n.conf[`$is`]
	name, ok := n.conf[`$name`]
	if ok {
		m[`$name`] = name
	}
	perm, ok := n.conf[`$permission`]
	if ok && perm != nil && perm != "read" {
		m[`$permission`] = perm
	}
	if n.valType != "" {
		m[`$type`] = n.valType
	}

	// TODO: Check for: invokable, type and interface

	return m
}

func (n *SimpleNode) GetType() dslink.ValueType {
	return n.valType
}

func (n *SimpleNode) SetType(t dslink.ValueType) {
	n.conf[`$type`] = t
	n.valType = t
}

func (n *SimpleNode) UpdateValue(v interface{}) {
	n.value = v
	// TODO: Something about the subscription and stuff
	val := dslink.NewValueUpdate(v)
	n.notifySubs(val)
}

func (n *SimpleNode) Value() interface{} {
	return n.value
}

func NewNode(name string, provider dslink.Provider) *SimpleNode {
	sn := &SimpleNode{
		name: name,
		p:    provider,
		attr: make(map[string]interface{}),
		conf: make(map[string]interface{}),
		chld: make(map[string]dslink.Node),
	}

	sn.conf[`$is`] = `node`

	return sn
}
