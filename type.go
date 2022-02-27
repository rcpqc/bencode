package bencode

import (
	"reflect"
	"sort"
	"strings"
	"sync"
)

type tyField struct {
	Name   string
	Option string
	Index  int
}

type tyProfile struct {
	Map   map[string]*tyField
	Order []*tyField
}

var tyProfiles sync.Map

func tyGet(rt reflect.Type) *tyProfile {
	profile, ok := tyProfiles.Load(rt)
	if !ok {
		profile = tyParse(rt)
		tyProfiles.Store(rt, profile)
	}
	return profile.(*tyProfile)
}

func tyParse(rt reflect.Type) *tyProfile {
	profile := &tyProfile{
		Order: make([]*tyField, 0, rt.NumField()),
		Map:   make(map[string]*tyField, rt.NumField()),
	}
	for i := 0; i < rt.NumField(); i++ {
		alias, opt := tagParse(rt.Field(i).Tag.Get("bencode"))
		if alias == "-" {
			continue
		}
		if alias == "" {
			alias = rt.Field(i).Name
		}
		field := &tyField{alias, opt, i}
		profile.Order = append(profile.Order, field)
		profile.Map[alias] = field
	}
	sort.Slice(profile.Order, func(i, j int) bool { return profile.Order[i].Name < profile.Order[j].Name })
	return profile
}

func tagParse(tag string) (alias string, opt string) {
	tags := strings.Split(tag, ",")
	alias = tags[0]
	if len(tags) > 1 {
		opt = tags[1]
	}
	return
}
