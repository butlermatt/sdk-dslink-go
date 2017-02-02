package dslink

type Provider interface {
	// TODO
	GetNode(path string) (Node, bool)
	GetRoot() Node
	//GetOrCreateNode(path string) Node;
}