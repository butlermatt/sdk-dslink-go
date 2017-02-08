package nodes

import (
	"errors"
	"fmt"
	"github.com/butlermatt/dslink"
	"strings"
)

// ValueType represents the type of value stored by the Node
type ValueType string

const (
	// ValueBool indicates this value type is a boolean
	ValueBool ValueType = "bool"
	// ValueNum indicates this value type is a number (integer or double)
	ValueNum ValueType = "num"
	// ValueString indicates this value type is a String
	ValueString ValueType = "string"
	// ValueDynamic indicates this value type is of an undetermined type
	ValueDynamic ValueType = "dynamic"
	// ValueDynamic indicates this value type is a Map
	ValueMap ValueType = "map"
	// ValueDynamic indicates this value type is an Array
	ValueArray ValueType = "array"
)

func GenerateEnumValue(options ...string) ValueType {
	return ValueType(fmt.Sprintf("enum[%s]", strings.Join(options, ",")))
}

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
	return nil
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

func (n *SimpleNode) RemoveChild(name string) error {
	return nil
}

func (n *SimpleNode) RemoveNode(node dslink.Node) error {
	return nil
}

func (n *SimpleNode) notifyList(name string, value interface{}) {
	dslink.Log.Printf("There are %d nodes to notify\n", len(n.listSubs))
	for _, i := range n.listSubs {
		r := &dslink.Response{Rid: i}
		r.AddUpdate(name, value)
		n.p.SendResponse(r)
	}
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
		dslink.Log.Printf("Closed link for Rid: %d", request.Rid)
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
	// TODO: Check for: invokable, type and interface

	return m
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
