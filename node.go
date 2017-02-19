package dslink

import (
	"time"
)

type Node interface {
	Name() string
	Attributes() map[string]interface{}
	GetAttribute(string) (interface{}, bool)
	SetAttribute(string, interface{})
	Configs() map[NodeConfig]interface{}
	GetConfig(NodeConfig) (interface{}, bool)
	SetConfig(NodeConfig, interface{})
	//Children() map[string]Node
	//AddChild(Node) error
	//RemoveChild(string) Node
	//Remove()
	//GetChild(string) Node
}

type Mapper interface {
	ToMap() map[string]interface{}
}

type Lister interface {
	List(*Request) *Response
	Close(*Request)
}

type Valued interface {
	GetType() ValueType
	SetType(ValueType)
	UpdateValue(interface{})
	Value() interface{}
	Subscribe(int32)
	Unsubscribe(int32)
}

//
type OnSetValue func(Node, interface{}) bool

type Settable interface {
	Set(*Request) *MsgErr
	EnableSet(PermType, OnSetValue)
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
