package bencode

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"
)

var (
	decoders = []func(buf *bytes.Buffer, rv reflect.Value) (err error){}
)

func init() {
	decoders = make([]func(buf *bytes.Buffer, rv reflect.Value) (err error), 256)
	decoders['i'] = intDecoder
	decoders['l'] = listDecoder
	decoders['d'] = dictDecoder
	decoders['0'] = strDecoder
	decoders['1'] = strDecoder
	decoders['2'] = strDecoder
	decoders['3'] = strDecoder
	decoders['4'] = strDecoder
	decoders['5'] = strDecoder
	decoders['6'] = strDecoder
	decoders['7'] = strDecoder
	decoders['8'] = strDecoder
	decoders['9'] = strDecoder
}

func decode(buf *bytes.Buffer, rv reflect.Value) error {
	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	decoder := decoders[b]
	if decoder == nil {
		return fmt.Errorf("decoder(%s) not supported", string(b))
	}
	buf.UnreadByte()
	return decoder(buf, indirect(rv))
}

func indirect(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	return rv
}

func readInteger(buf *bytes.Buffer) (int64, error) {
	var integer int64 = 0
	minus := false
	for length := 0; true; length++ {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("readInteger ReadByte err: %v", err)
		}
		if b == '-' && length == 0 {
			minus = true
			continue
		}
		if b < '0' || b > '9' {
			buf.UnreadByte()
			break
		}
		if 10*integer+int64(b-'0') < integer {
			return 0, fmt.Errorf("intDecoder integer overflow")
		}
		integer = 10*integer + int64(b-'0')
	}
	if minus {
		if integer == 0 {
			return 0, fmt.Errorf("minus zero is illegal")
		}
		return -integer, nil
	}
	return integer, nil
}

func readAssert(buf *bytes.Buffer, expected byte) error {
	if b, err := buf.ReadByte(); err != nil {
		return err
	} else if expected != b {
		return fmt.Errorf("expect(%d) got(%d)", expected, b)
	}
	return nil
}

func intDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	if err := readAssert(buf, 'i'); err != nil {
		return err
	}
	integer, err := readInteger(buf)
	if err != nil {
		return err
	}
	if err := readAssert(buf, 'e'); err != nil {
		return err
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv.SetInt(integer)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv.SetUint(uint64(integer))
	case reflect.Bool:
		rv.SetBool(integer != 0)
	case reflect.Interface:
		rv.Set(reflect.ValueOf(int(integer)))
	default:
		return fmt.Errorf("value's kind(%v) not match", rv.Kind())
	}
	return nil
}

func strDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	length, err := readInteger(buf)
	if err != nil {
		return err
	}
	if length < 0 {
		return fmt.Errorf("strDecoder parse error")
	}
	if err := readAssert(buf, ':'); err != nil {
		return err
	}
	if buf.Len() < int(length) {
		return fmt.Errorf("strDecoder parse error")
	}
	bytes := buf.Next(int(length))
	str := *(*string)(unsafe.Pointer(&bytes))
	if rv.Kind() == reflect.String {
		rv.SetString(str)
	} else if rv.Kind() == reflect.Interface {
		rv.Set(reflect.ValueOf(str))
	} else {
		return fmt.Errorf("value's kind(%v) not match", rv.Kind())
	}
	return nil
}

func listDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	if err := readAssert(buf, 'l'); err != nil {
		return err
	}
	if rv.Kind() == reflect.Slice {
		return listSliceDecoder(buf, rv)
	} else if rv.Kind() == reflect.Interface {
		sli := []interface{}{}
		rsli := reflect.ValueOf(&sli)
		if err := listSliceDecoder(buf, rsli.Elem()); err != nil {
			return err
		}
		rv.Set(rsli.Elem())
	}
	return nil
}

func listSliceDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("listSliceDecoder ReadByte err: %v", err)
		}
		if b == 'e' {
			break
		}
		buf.UnreadByte()

		// Read Value
		rval := reflect.New(rv.Type().Elem())
		if err := decode(buf, rval.Elem()); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, rval.Elem()))
	}
	return nil
}

func dictDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	if err := readAssert(buf, 'd'); err != nil {
		return err
	}
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		return dictMapDecoder(buf, rv)
	} else if rv.Kind() == reflect.Struct {
		return dictStructDecoder(buf, rv)
	} else if rv.Kind() == reflect.Invalid {
		rmap := reflect.ValueOf(map[string]interface{}{})
		if err := dictMapDecoder(buf, rmap); err != nil {
			return err
		}
		rv.Set(rmap)
	}
	return fmt.Errorf("type not match")
}

func dictMapDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("dictMapDecoder ReadByte err: %v", err)
		}
		if b == 'e' {
			break
		}
		buf.UnreadByte()
		// Read Key
		key := ""
		rkey := reflect.ValueOf(&key)
		if err := strDecoder(buf, rkey.Elem()); err != nil {
			return err
		}
		// Read Value
		rval := reflect.New(rv.Type().Elem())
		if err := decode(buf, rval.Elem()); err != nil {
			return err
		}
		rv.SetMapIndex(rkey.Elem(), rval.Elem())
	}
	return nil
}

func dictStructDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("dictMapDecoder ReadByte err: %v", err)
		}
		if b == 'e' {
			break
		}
		buf.UnreadByte()
		// Read Key
		key := ""
		rkey := reflect.ValueOf(&key)
		if err := strDecoder(buf, rkey.Elem()); err != nil {
			return err
		}
		// Read Value
		rval := tagField(rv, key)
		if !rval.IsValid() {
			var val interface{}
			rval = reflect.ValueOf(&val).Elem()
		}
		if err := decode(buf, rval); err != nil {
			return err
		}
	}
	return nil
}
