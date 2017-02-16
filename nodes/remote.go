package nodes

import (
	"sync"
	"github.com/butlermatt/dslink"
	"errors"
)

type RemoteNode struct {
	name string
	aMu  sync.RWMutex
	attr map[string]interface{}
	cMu  sync.RWMutex
	conf map[dslink.NodeConfig]interface{}
	cdMu sync.RWMutex
	chdn map[string]*RemoteNode
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

func (n *RemoteNode) RemoveChild(p string) dslink.Node {
	n.cdMu.Lock()
	defer n.cdMu.Unlock()

	nd := n.chdn[p]
	delete(n.chdn, p)

	return nd
}

func (n *RemoteNode) Remove() {
	// TODO Not supported? Or send request to remote to execute it?
}

func (n *RemoteNode) GetChild(p string) dslink.Node {
	n.cdMu.RLock()
	defer n.cdMu.RUnlock()

	return n.chdn[p]
}
