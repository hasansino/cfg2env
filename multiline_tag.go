package cfg2env

import (
	"strconv"
	"strings"
)

// StructTag is a copy of original golang source code
// with implemented support for multi-line struct tags.
type StructTag string

func (tag StructTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

func (tag StructTag) Lookup(key string) (value string, ok bool) {
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && validChar(tag[i]) && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}

		// Remove CR, LF, tab and space chars from name
		name := sanitizeName(string(tag[:i]))
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}

		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			// A workaround for strconv.Unquote, which does not accept
			// newline (\n) chars and return error.
			value, err := strconv.Unquote(escapeValue(qvalue))
			if err != nil {
				break
			}
			return value, true
		}
	}
	return "", false
}

func validChar(r uint8) bool {
	switch {
	case r == 0x0A || r == 0x0D: // LF and CR
		return true
	case r == 0x09: // horizontal tab
		return true
	case r >= ' ':
		return true
	default:
		return false
	}
}

func sanitizeName(s string) string {
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func escapeValue(s string) string {
	return strconv.Quote(strings.Trim(s, `"`))
}
