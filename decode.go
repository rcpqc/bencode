package bencode

import (
	"bytes"
	"fmt"
	"reflect"
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
		return fmt.Errorf("decoder(%d) not supported", b)
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
	if rv.Kind() == reflect.Map {
		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rv.Type()))
		}
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
	if rv.Kind() == reflect.String {
		rv.SetString(string(bytes))
	} else if rv.Kind() == reflect.Interface {
		rv.Set(reflect.ValueOf(string(bytes)))
	}
	return nil
}

func makeElement(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map {
		return reflect.New(rv.Type().Elem()).Elem()
	} else if rv.Kind() == reflect.Interface {
		var elem interface{}
		return reflect.ValueOf(&elem).Elem()
	}
	return reflect.Value{}
}

func listDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	rsli := reflect.Value{}
	if rv.Kind() == reflect.Slice {
		rsli = rv
	} else if rv.Kind() == reflect.Interface {
		sli := []interface{}{}
		rsli = reflect.ValueOf(&sli).Elem()
	}

	if err := readAssert(buf, 'l'); err != nil {
		return err
	}
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("listDecoder ReadByte err: %v", err)
		}
		if b == 'e' {
			break
		}
		buf.UnreadByte()

		// Read Value
		relem := makeElement(rv)
		if err := decode(buf, relem); err != nil {
			return err
		}
		if rsli.Kind() == reflect.Slice {
			rsli.Set(reflect.Append(rsli, relem))
		}
	}

	if rsli.IsValid() {
		rv.Set(rsli)
	}

	return nil
}

func dictDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	if rv.Kind() == reflect.Struct {
		return dictStructDecoder(buf, rv)
	} else {
		return dictMapDecoder(buf, rv)
	}
}

func dictMapDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	rmap := reflect.Value{}
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		rmap = rv
	} else if rv.Kind() == reflect.Interface {
		m := map[string]interface{}{}
		rmap = reflect.ValueOf(&m).Elem()
	}
	if err := readAssert(buf, 'd'); err != nil {
		return err
	}
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
		rkey := reflect.ValueOf(&key).Elem()
		if err := strDecoder(buf, rkey); err != nil {
			return err
		}
		// Read Value
		rval := makeElement(rv)
		if err := decode(buf, rval); err != nil {
			return err
		}
		if rmap.Kind() == reflect.Map {
			rmap.SetMapIndex(rkey, rval)
		}
	}

	if rmap.IsValid() {
		rv.Set(rmap)
	}
	return nil
}

func dictStructDecoder(buf *bytes.Buffer, rv reflect.Value) error {
	fields := tyGet(rv.Type()).Map
	if err := readAssert(buf, 'd'); err != nil {
		return err
	}
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("dictStructDecoder ReadByte err: %v", err)
		}
		if b == 'e' {
			break
		}
		buf.UnreadByte()
		// Read Key
		key := ""
		rkey := reflect.ValueOf(&key).Elem()
		if err := strDecoder(buf, rkey); err != nil {
			return err
		}

		// Find Field
		rfield := reflect.Value{}
		if field, ok := fields[key]; ok {
			rfield = rv.Field(field.Index)
		}

		// Read Value
		if err := decode(buf, rfield); err != nil {
			return err
		}
	}
	return nil
}
