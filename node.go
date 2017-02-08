package dslink

import (
	"fmt"
	"strings"
	"time"
)

type Node interface {
	// TODO
	GetAttribute(string) (interface{}, bool)
	GetConfig(NodeConfig) (interface{}, bool)
	AddChild(Node) error
	RemoveChild(string) Node
	Remove()
	GetChild(string) Node
	List(*Request) *Response
	Close(*Request)
	ToMap() map[string]interface{}
}

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

type Valued interface {
	GetType() ValueType
	SetType(ValueType)
	UpdateValue(interface{})
	Value() interface{}
	Subscribe(int32)
	Unsubscribe(int32)
}

type ValueEditor interface {
	Valued
	GetEditor() string
	SetEditor(string)
}

type Invokable interface {
	Invoke(map[string]interface{})
}

type ValueUpdate struct {
	value interface{}
	ts    time.Time
}

func (v *ValueUpdate) GetTs() time.Time {
	return v.ts
}

func (v *ValueUpdate) Value() interface{} {
	return v.value
}

func NewValueUpdate(value interface{}) *ValueUpdate {
	return &ValueUpdate{value: value, ts: time.Now()}
}