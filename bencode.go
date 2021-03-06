package bencode

import (
	"bytes"
	"reflect"
)

// Marshal
func Marshal(v interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	if err := encode(&buf, reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal
func Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	if err := decode(buf, reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}
