package nodes

import (
	"sync"
	"errors"
	"github.com/butlermatt/dslink"
	//"github.com/butlermatt/dslink/log"
)

type RemoteNode struct {
	name string
	path string
	aMu  sync.RWMutex
	attr map[string]interface{}
	cMu  sync.RWMutex
	conf map[dslink.NodeConfig]interface{}
	cdMu sync.RWMutex
	chdn map[string]*RemoteNode
}

func (n *RemoteNode) Name() string {
	return n.name
}

func (n *RemoteNode) Path() string {
	return n.path
}

func (n *RemoteNode) Attributes() map[string]interface{} {
	return n.attr
}

func (n *RemoteNode) GetAttribute(name string) (interface{}, bool) {
	n.aMu.RLock()
	defer n.aMu.RUnlock()

	v, ok := n.attr[name]
	return v, ok
}

func (n *RemoteNode) SetAttribute(name string, v interface{}) {
	n.aMu.Lock()
	defer n.aMu.Unlock()
	n.attr[name] = v
}

func (n *RemoteNode) Configs() map[dslink.NodeConfig]interface{} {
	return n.conf
}

func (n *RemoteNode) GetConfig(c dslink.NodeConfig) (interface{}, bool) {
	n.cMu.RLock()
	defer n.cMu.RUnlock()

	v, ok := n.conf[c]
	return v, ok
}

func (n *RemoteNode) SetConfig(c dslink.NodeConfig, v interface{}) {
	n.cMu.Lock()
	defer n.cMu.Unlock()

	n.conf[c] = v
}

func (n *RemoteNode) Children() map[string]*RemoteNode {
	return n.chdn
}

func (n *RemoteNode) AddChild(node dslink.Node) error {
	rn, ok := node.(*RemoteNode)

	if !ok {
		return errors.New("Specified Node is not a valid child for a remote node")
	}

	n.cdMu.Lock()
	defer n.cdMu.Unlock()

	n.chdn[rn.name] = rn
	return nil
}

func (n *RemoteNode) RemoveChild(p string) *RemoteNode {
	n.cdMu.Lock()
	defer n.cdMu.Unlock()

	nd := n.chdn[p]
	delete(n.chdn, p)

	return nd
}

func (n *RemoteNode) Remove() {
	// TODO Not supported? Or send request to remote to execute it?
}

func (n *RemoteNode) GetChild(p string) *RemoteNode {
	n.cdMu.RLock()
	defer n.cdMu.RUnlock()

	return n.chdn[p]
}

func (n *RemoteNode) Type() dslink.ValueType {
	t := n.conf[dslink.ConfigType]
	if t == nil {
		return ""
	}
	vt, ok := t.(string)
	if !ok {
		return ""
	}
	return dslink.ValueType(vt)
}

func NewRemoteNode(p string) *RemoteNode {
	r := &RemoteNode{
		name: PathName(p),
		path: p,
		conf: make(map[dslink.NodeConfig]interface{}),
		attr: make(map[string]interface{}),
		chdn: make(map[string]*RemoteNode),
	}
	return r
}

func NewRemoteFromMap(n, p string, m map[interface{}]interface{}) *RemoteNode {
	r := NewRemoteNode(n)
	var pa string
	if p == "/" || (len(p) > 1 && p[len(p) - 1] == '/') {
		pa = p + n
	} else {
		pa = p + "/" + n
	}
	r.path = pa
	for k, v := range m {
		nm, _ := k.(string)
		switch nm[0] {
		case '$':
			r.SetConfig(dslink.NodeConfig(nm), v)
		case '@':
			r.SetAttribute(nm, v)
		}
	}

	return r
}