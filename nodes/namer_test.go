package nodes

import "testing"

var cases = []struct {
	decoded, encoded string
}{
	{``, ``},
	{`Hello, World`, `Hello, World`},
	{`This\is\a\test`, `This%5Cis%5Ca%5Ctest`},
	{`Don't%20Escape%20Escapes`, `Don't%20Escape%20Escapes`},
	{`%./\?*:|<>$@`, `%25%2E%2F%5C%3F%2A%3A%7C%3C%3E%24%40`},
	{`%25%2E%2F%5C%3F%2A%3A%7C%3C%3E%24%40`, `%25%2E%2F%5C%3F%2A%3A%7C%3C%3E%24%40`},
	{`Hello, 世界`, `Hello, 世界`},
}

func TestEncodeName(t *testing.T) {
	for _, c := range cases {
		got := EncodeName(c.decoded)
		if got != c.encoded {
			t.Errorf("EncodeName(%q) == %q, want %q", c.decoded, got, c.encoded)
		}
	}
}

func TestDecodeName(t *testing.T) {
	for _, c := range cases {
		got := DecodeName(c.encoded)
		if got != c.decoded {
			t.Errorf("DecodeName(%q) == %q, want %q", c.encoded, got, c.decoded)
		}
	}
}
