package provider

import (
	"github.com/butlermatt/dslink"
	"github.com/butlermatt/dslink/nodes"
)

type SimpleProvider struct {
	cache map[string]dslink.Node
	root dslink.Node
}

func (s *SimpleProvider) GetNode(path string) (dslink.Node, bool) {
	nd, ok := s.cache[path]
	return nd, ok
}

func (s *SimpleProvider) GetRoot() dslink.Node {
	return s.root
}

func New() *SimpleProvider {
	r := nodes.New()
	sp := &SimpleProvider{cache: make(map[string]dslink.Node), root: r}
	sp.cache["/"] = r
	return sp
}