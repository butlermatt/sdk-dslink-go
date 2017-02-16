package nodes_test

import (
	"testing"
	"github.com/butlermatt/dslink/nodes"
)

func TestPathName(t *testing.T) {
	var cases = []struct {
		path, name string
	}{
		{``, ``},
		{`/`, ``},
		{`///`, ``},
		{`/Hello`, `Hello`},
		{`/Hello/There`, `There`},
		{`/Hello/There/`, `There`},
	}
	for _, c := range cases {
		got := nodes.PathName(c.path)
		if got != c.name {
			t.Errorf("PathName(%q) == %q, want %q", c.path, got, c.name)
		}
	}
}