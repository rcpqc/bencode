package bencode

import (
	"reflect"
	"sort"
	"strings"
)

const TAG_KEY = "bencode"

type Field struct {
	Name   string
	Option string
	Index  int
}

func tyParseSlice(rt reflect.Type) []Field {
	fields := make([]Field, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		alias, opt := tagParse(rt.Field(i).Tag.Get(TAG_KEY))
		if alias == "-" {
			continue
		}
		if alias == "" {
			alias = rt.Field(i).Name
		}
		fields = append(fields, Field{alias, opt, i})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

func tyParseMap(rt reflect.Type) map[string]Field {
	fields := make(map[string]Field, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		alias, opt := tagParse(rt.Field(i).Tag.Get(TAG_KEY))
		if alias == "-" {
			continue
		}
		if alias == "" {
			alias = rt.Field(i).Name
		}
		fields[alias] = Field{alias, opt, i}
	}
	return fields
}

func tagParse(tag string) (alias string, opt string) {
	tags := strings.Split(tag, ",")
	alias = tags[0]
	if len(tags) > 1 {
		opt = tags[1]
	}
	return
}
