package nodes

import (
	"strings"
	"bytes"
	"fmt"
)

// BannedChars are characters which may not appear in a node path.
var BannedChars = [...]rune{
	'%',
	'.',
	'/',
	'\\',
	'?',
	'*',
	':',
	'|',
	'<',
	'>',
	'$',
	'@',
}

// CreateName accepts a string input and will automatically escape any invalid
// characters returning a string which can be used as part of a node path.
func CreateName(in string) string {
	r := strings.NewReader(in)
	buf := new(bytes.Buffer)
	var err error
	var c rune
	MAIN_LOOP:
	for err == nil {
		c, _, err = r.ReadRune()
		if err != nil {
			break
		}
		// Check for already escaped characters.
		if c == '%' {
			n, _, err := r.ReadRune()
			if err != nil {
				buf.WriteString(fmt.Sprintf("%%%X", c))
				break
			}
			if isHex(n) {
				buf.WriteRune(c)
				buf.WriteRune(n)

				n, _, err = r.ReadRune()
				if err != nil {
					break
				}
				if isHex(n) {
					buf.WriteRune(n)
					continue
				} else {
					r.UnreadRune()
					continue
				}
			} else {
				r.UnreadRune()
			}
		}
		for _, b := range BannedChars {
			if c == b {
				buf.WriteString(fmt.Sprintf("%%%X", c))
				continue MAIN_LOOP
			}
		}

		buf.WriteRune(c)
	}

	return buf.String()
}

func isHex(c rune) bool {
	if c >= '0' && c <= '9' {
		return true
	}
	if c >= 'a' && c <= 'f' {
		return true
	}
	if c >= 'A' && c <= 'F' {
		return true
	}
	return false
}