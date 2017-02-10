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

type Column struct {
	Name    string
	Type    ValueType
	Default interface{}
}

type Params map[ParamVal]interface{}

func (p Params) Get(val ParamVal) interface{} {
	return p[val]
}

func (p Params) Set(key ParamVal, val interface{}) {
	p[key] = val
}

type InvokeFn func(map[string]interface{}, chan<-[]interface{})

type Invokable interface {
	Invoke(*Request)
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
