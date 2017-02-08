package dslink

import (
	lg "log"
)

type PermType string

const (
	PermNone   PermType = "none"
	PermRead   PermType = "read"
	PermWrite  PermType = "write"
	PermConfig PermType = "config"
	PermNever  PermType = "never"
)

type NodeConfig string

const (
	ConfigBase         NodeConfig = "$base"
	ConfigIs           NodeConfig = "$is"
	ConfigInterface    NodeConfig = "$interface"
	ConfigDisconnected NodeConfig = "$disconnectedTs"
	ConfigPermission   NodeConfig = "$permission"
	ConfigPermissions  NodeConfig = "$$permissions"
	ConfigName         NodeConfig = "$name"
	ConfigType         NodeConfig = "$type"
	ConfigWritable     NodeConfig = "$writable"
	ConfigSetings      NodeConfig = "$settings"
	ConfigParams       NodeConfig = "$params"
	ConfigColumns      NodeConfig = "$columns"
	ConfigResult       NodeConfig = "$result"
	ConfigStreamMeta   NodeConfig = "$streamMeta"
	ConfigInvokable    NodeConfig = "$invokable"
)

type ParamVal string

const (
	ParamName   ParamVal = "name"
	ParamType   ParamVal = "type"
	ParamDef    ParamVal = "default"
	ParamEditor ParamVal = "editor"
	ParamPlace  ParamVal = "placeholder"
	ParamDesc   ParamVal = "description"
)

var Log *lg.Logger

type Provider interface {
	// TODO
	GetNode(path string) (Node, bool)
	GetRoot() Node
	HandleRequest(req *Request) *Response
	SendResponse(resp *Response)
	AddNode(path string, node Node)
	RemoveNode(path string) Node
	//GetOrCreateNode(path string) Node;
}
