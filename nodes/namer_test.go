package nodes

import "testing"

var cases = []struct {
	decoded, encoded string
}{
	{``, ``},
	{`Hello, World`, `Hello, World`},
	{`This\is\a\test`, `This%5Cis%5Ca%5Ctest`},
	{`%/\?*:|<>$@'"`, `%25%2F%5C%3F%2A%3A%7C%3C%3E%24%40%27%22`},
	{`Hello, 世界`, `Hello, 世界`},
	{`Dots.allowed`, `Dots.allowed`},
}

func TestEncodeName(t *testing.T) {
	encodeCases := append(cases, []struct {
		decoded, encoded string
	}{
		{`Don%27t%20Escape%20Escapes`, `Don%27t%20Escape%20Escapes`},
		{`%25%2E%2F%5C%3F%2A%3A%7C%3C%3E%24%40`, `%25%2E%2F%5C%3F%2A%3A%7C%3C%3E%24%40`},
	}...)

	for _, c := range encodeCases {
		got := EncodeName(c.decoded)
		if got != c.encoded {
			t.Errorf("EncodeName(%q) == %q, want %q", c.decoded, got, c.encoded)
		}
	}
}

func TestDecodeName(t *testing.T) {
	for _, c := range cases {
		got, err := DecodeName(c.encoded)
		if err != nil {
			t.Errorf("DecodeName(%q) == %q, want %q", c.encoded, err, c.decoded)
		}
		if got != c.decoded {
			t.Errorf("DecodeName(%q) == %q, want %q", c.encoded, got, c.decoded)
		}

	}

	var failCases = []string{`%`, `%AZ`}
	for _, c := range failCases {
		got, err := DecodeName(c)
		if err == nil {
			t.Errorf("DecodeName(%q) == %q. Should have errored", c, got)
		}
		if got != "" {
			t.Errorf("DecodeName(%q) == %q. Should have errored", c, got)
		}
	}
}
