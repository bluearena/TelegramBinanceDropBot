package main

import (
	"bytes"
	"strings"
	"strconv"
)

func Spacef(v float64) string {
	buf := &bytes.Buffer{}
	if v < 0 {
		buf.Write([]byte{'-'})
		v = 0 - v
	}

	space := []byte{','}

	parts := strings.Split(strconv.FormatFloat(v, 'f', 1, 64), ".")
	pos := 0
	if len(parts[0])%3 != 0 {
		pos += len(parts[0]) % 3
		buf.WriteString(parts[0][:pos])
		buf.Write(space)
	}
	for ; pos < len(parts[0]); pos += 3 {
		buf.WriteString(parts[0][pos: pos+3])
		buf.Write(space)
	}
	buf.Truncate(buf.Len() - 1)

	if len(parts) > 1 {
		buf.Write([]byte{'.'})
		buf.WriteString(parts[1])
	}
	return buf.String()
}