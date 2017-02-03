package nodes

import (
	"github.com/butlermatt/dslink"
	"fmt"
	"strings"
	"errors"
)

// ValueType represents the type of value stored by the Node
type ValueType string;

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
	provider dslink.Provider
	attr map[string]interface{}
	conf map[string]interface{}
	chld map[string]*SimpleNode
	Parent *SimpleNode
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

	return nil
}

func (n *SimpleNode) RemoveChild(name string) error {
	return nil
}

func (n *SimpleNode) RemoveNode(node dslink.Node) error {
	return nil
}

func New() *SimpleNode {
	return &SimpleNode{}
}