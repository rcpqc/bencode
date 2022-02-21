package bencode

import (
	"reflect"
	"strings"
)

func tagField(rv reflect.Value, tag string) reflect.Value {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		alias, _ := tagParse(rt.Field(i).Tag.Get("bencode"))
		if alias == "-" {
			continue
		}
		if alias == tag || rt.Field(i).Name == tag {
			return rv.Field(i)
		}
	}
	return reflect.Value{}
}

func tagParse(tag string) (alias string, opt string) {
	tags := strings.Split(tag, ",")
	alias = tags[0]
	if len(tags) > 1 {
		opt = tags[1]
	}
	return
}
