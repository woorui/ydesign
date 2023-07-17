package metadata

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

var (
	ErrMDKeyOrValue  = errors.New("metadata: key or value cannot contain '\n' and  ':'")
	ErrInvalidFormat = errors.New("metadata: invalid format")
)

type MD map[string]string

func (md MD) Get(key string) (string, bool) {
	value, ok := md[key]
	return value, ok
}

func (md MD) Set(key, value string) error {
	if strings.ContainsAny(key, ":\n") || strings.ContainsAny(value, ":\n") {
		return ErrMDKeyOrValue
	}
	md[key] = value
	return nil
}

func (md MD) Del(key string) {
	delete(md, key)
}

func (md MD) Clone() MD {
	if md == nil {
		return nil
	}
	if len(md) == 0 {
		return make(MD)
	}

	md2 := make(MD, len(md))
	for k, v := range md {
		md2[k] = v
	}

	return md2
}

func (md MD) Encode() ([]byte, error) {
	var (
		n   = 0
		buf = new(bytes.Buffer)
	)

	for k, v := range md {
		if n != len(md)-1 {
			buf.Grow(len(k) + len(v) + 1)
		} else {
			buf.Grow(len(k) + len(v) + 2)
		}
		buf.WriteString(k)
		buf.WriteString(":")
		buf.WriteString(v)

		if n != len(md)-1 {
			buf.WriteByte('\n')
		}
		n++
	}

	return buf.Bytes(), nil
}

func (md MD) Decode(data []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Bytes()

		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			return ErrInvalidFormat
		}

		md[string(parts[0])] = string(parts[1])
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}
