package dslink

type Node interface {
	// TODO
	GetAttribute(string) (interface{}, bool)
	GetConfig(string) (interface{}, bool)
	AddChild(Node) error
	RemoveChild(string) error
	RemoveNode(Node) error
	GetChild(string) Node
	List(*Request) *Response
	Close(*Request)
	ToMap() map[string]interface{}
}

type LocalNode interface {
	Node
	UpdateValue(interface{}, bool)
	GetValue() interface{}
}

type Invokable interface {
	Invoke(map[string]interface{})
}

type Subscriber interface {
	Subscribe()
	Unsubscribe()
}