package bencode

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

var (
	encoders = []func(buf *bytes.Buffer, rv reflect.Value) (err error){}
)

func init() {
	encoders = make([]func(buf *bytes.Buffer, rv reflect.Value) (err error), 32)
	encoders[reflect.Bool] = boolEncoder
	encoders[reflect.Int] = intEncoder
	encoders[reflect.Int8] = intEncoder
	encoders[reflect.Int16] = intEncoder
	encoders[reflect.Int32] = intEncoder
	encoders[reflect.Int64] = intEncoder
	encoders[reflect.Uint] = uintEncoder
	encoders[reflect.Uint8] = uintEncoder
	encoders[reflect.Uint16] = uintEncoder
	encoders[reflect.Uint32] = uintEncoder
	encoders[reflect.Uint64] = uintEncoder
	encoders[reflect.Array] = arrayEncoder
	encoders[reflect.Interface] = elemEncoder
	encoders[reflect.Map] = mapEncoder
	encoders[reflect.Ptr] = elemEncoder
	encoders[reflect.Slice] = arrayEncoder
	encoders[reflect.String] = strEncoder
	encoders[reflect.Struct] = structEncoder
}

func encode(buf *bytes.Buffer, rv reflect.Value) error {
	encoder := encoders[rv.Kind()]
	if encoder == nil {
		return fmt.Errorf("encoder(%s) not supported", rv.Kind().String())
	}
	return encoder(buf, rv)
}

func boolEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	if rv.Bool() {
		_, err := buf.Write([]byte{'i', '1', 'e'})
		return err
	}
	_, err := buf.Write([]byte{'i', '0', 'e'})
	return err
}

func intEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	buf.Write([]byte{'i'})
	var dst [64]byte
	buf.Write(strconv.AppendInt(dst[:0], rv.Int(), 10))
	buf.Write([]byte{'e'})
	return nil
}

func uintEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	buf.Write([]byte{'i'})
	var dst [64]byte
	buf.Write(strconv.AppendUint(dst[:0], rv.Uint(), 10))
	buf.Write([]byte{'e'})
	return nil
}

func arrayEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	buf.Write([]byte{'l'})
	for i := 0; i < rv.Len(); i++ {
		if err := encode(buf, rv.Index(i)); err != nil {
			return err
		}
	}
	buf.Write([]byte{'e'})
	return nil
}

func elemEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	return encode(buf, rv.Elem())
}

func mapEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	buf.Write([]byte{'d'})
	if rv.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("key of map must be string")
	}
	keys := rv.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })

	for _, key := range keys {
		val := rv.MapIndex(key)
		if err := strEncoder(buf, key); err != nil {
			return err
		}
		if err := encode(buf, val); err != nil {
			return err
		}
	}
	buf.Write([]byte{'e'})
	return nil
}

func strEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	str := rv.String()
	length := int64(len(str))
	var dst [64]byte
	buf.Write(strconv.AppendInt(dst[:0], length, 10))
	buf.Write([]byte{':'})
	buf.WriteString(str)
	return nil
}

func structEncoder(buf *bytes.Buffer, rv reflect.Value) error {
	buf.Write([]byte{'d'})
	for _, field := range tyGet(rv.Type()).Order {
		if field.Option == "omitempty" && rv.Field(field.Index).IsZero() {
			continue
		}
		if err := strEncoder(buf, reflect.ValueOf(field.Name)); err != nil {
			return err
		}
		if err := encode(buf, rv.Field(field.Index)); err != nil {
			return err
		}
	}
	buf.Write([]byte{'e'})
	return nil
}
