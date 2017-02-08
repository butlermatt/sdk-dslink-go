package dslink

type Node interface {
	// TODO
	GetAttribute(string) (interface{}, bool)
	GetConfig(string) (interface{}, bool)
	AddChild(Node) error
	RemoveChild(string) Node
	Remove()
	GetChild(string) Node
	List(*Request) *Response
	Close(*Request)
	ToMap() map[string]interface{}
}

type Valued interface {
	GetType() string
	SetType(string)
	UpdateValue(interface{})
	Value() interface{}
}

type ValueEditor interface {
	Valued
	GetEditor() string
	SetEditor(string)
}

type Invokable interface {
	Invoke(map[string]interface{})
}

type Subscriber interface {
	Subscribe()
	Unsubscribe()
}