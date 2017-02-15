package dslink

import (
	"fmt"
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

const (
	ResultValues = "values"
	ResultTable  = "table"
	ResultStream = "stream"
)

// PermType is the Permission type.
type PermType string

func (p PermType) Level() int {
	switch p {
	case PermNone:
		return 0
	case PermRead:
		return 1
	case PermWrite:
		return 2
	case PermConfig:
		return 3
	case PermNever:
		return 4
	default:
		return -1
	}
}

const (
	// PermNone is permission none.
	PermNone   PermType = "none"
	// PermRead is permission Read only
	PermRead   PermType = "read"
	// PermWrite is permission write
	PermWrite  PermType = "write"
	// PermConfig is permission config
	PermConfig PermType = "config"
	// PermNever is permission never
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
	// ParamName is the name of the parameter. Used in Invoke actions.
	ParamName   ParamVal = "name"
	// ParamType is the type of the parameter. Used in Invoke actions.
	ParamType   ParamVal = "type"
	// ParamDef is the default value of the parameter. Used in Invoke actions.
	ParamDef    ParamVal = "default"
	// ParamEditor is the editor for the type. Used in Invoke actions.
	ParamEditor ParamVal = "editor"
	// ParamHolder is the placeholder for the value. Used in Invoke actions.
	ParamHolder ParamVal = "placeholder"
	// ParamDesc is the parameter description. Used in Invoke actions.
	ParamDesc   ParamVal = "description"
)

type MethodType string;
const (
	MethodList    MethodType = "list"
	MethodSub     MethodType = "subscribe"
	MethodUnsub   MethodType = "unsubscribe"
	MethodClose   MethodType = "close"
	MethodSet     MethodType = "set"
	MethodRemove  MethodType = "remove"
	MethodInvoke  MethodType = "invoke"
)