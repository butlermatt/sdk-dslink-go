package nodes

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// BannedChars are characters which may not appear in a node path.
var BannedChars = [...]rune{
	'%',
	//'.', // No longer banned
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
	'\'',
	'"',
}

// EncodeName accepts a string input and will automatically escape any invalid
// characters returning a string which can be used as part of a node path.
func EncodeName(in string) string {
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

// DecodeName accepts an encoded name string and returns a decoded representation
// of the string.
func DecodeName(in string) (string, error) {
	r := strings.NewReader(in)
	buf := new(bytes.Buffer)
	var err error
	var c rune

	for err == nil {
		c, _, err = r.ReadRune()
		if err != nil {
			break
		}
		if c == '%' {
			// Read next two characters as hex and parse the number to a byte.
			var s string
			for i := 0; i < 2; i++ {
				n, _, err := r.ReadRune()

				if err != nil || !isHex(n) {
					return "", fmt.Errorf("unable to decode string: \"%q\"", in)
				}
				s += string(n)
			}

			b, err := strconv.ParseInt(s, 16, 0)
			if err != nil {
				return "", fmt.Errorf("unable to decode string: \"%q\"", in)
			}
			buf.WriteByte(byte(b))
			continue
		}
		buf.WriteRune(c)
	}

	return buf.String(), nil
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
