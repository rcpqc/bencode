package bencode

import (
	"fmt"
	"reflect"
	"testing"
)

type TestCase struct {
	Entity interface{}
	Data   []byte
}

type TestS1 struct {
	Sf   string `bencode:"sf"`
	Ffgd int    `bencode:"ffgd"`
	HHd  string `bencode:"hhd,omitempty"`
	XXYh uint32 `bencode:"-"`
}

type TestS2 struct {
	GGG   string
	IFace interface{} `bencode:"iface"`
}

type TestS3 struct {
	S1  *TestS1 `bencode:"s1"`
	SS3 bool
}

var comTests = []TestCase{
	{Entity: 23, Data: []byte("i23e")},
	{Entity: -422, Data: []byte("i-422e")},
	{Entity: -0, Data: []byte("i0e")},
	{Entity: uint32(99), Data: []byte("i99e")},
	{Entity: "abc", Data: []byte("3:abc")},
	{Entity: "", Data: []byte("0:")},
	{Entity: "3242434te", Data: []byte("9:3242434te")},
	{Entity: true, Data: []byte("i1e")},
	{Entity: false, Data: []byte("i0e")},
	{Entity: []int{1, 2, 3, 4, 5}, Data: []byte("li1ei2ei3ei4ei5ee")},
	{Entity: []string{"aa", "b", "ccc"}, Data: []byte("l2:aa1:b3:ccce")},
	{Entity: []interface{}{"aa", "b", 33, -23, "XX"}, Data: []byte("l2:aa1:bi33ei-23e2:XXe")},
	{Entity: map[string]int{"aa": 43, "bbbfe": -544, "": 0}, Data: []byte("d0:i0e2:aai43e5:bbbfei-544ee")},
	{Entity: map[string]interface{}{"a": "1", "b": "2", "c": 3, "d": 4, "e": "5"}, Data: []byte("d1:a1:11:b1:21:ci3e1:di4e1:e1:5e")},
	{Entity: TestS3{S1: &TestS1{Sf: "gjc", Ffgd: 87}, SS3: true}, Data: []byte("d3:SS3i1e2:s1d4:ffgdi87e2:sf3:gjcee")},
}

var encTests = []TestCase{
	{Entity: TestS1{"xxx", 2, "", 556}, Data: []byte("d4:ffgdi2e2:sf3:xxxe")},
	{Entity: TestS1{"xxx", 2, "66", 556}, Data: []byte("d4:ffgdi2e3:hhd2:662:sf3:xxxe")},
	{Entity: TestS2{GGG: "gggee", IFace: []int{4, 6}}, Data: []byte("d3:GGG5:gggee5:ifaceli4ei6eee")},
	{Entity: TestS2{GGG: "gggee", IFace: map[string]int{"gds": 8, "353": -45}}, Data: []byte("d3:GGG5:gggee5:ifaced3:353i-45e3:gdsi8eee")},
	{Entity: [3]int{1, 3, 5}, Data: []byte("li1ei3ei5ee")},
}

var decTests = []TestCase{
	{Entity: TestS1{"xxx", 2, "", 0}, Data: []byte("d4:ffgdi2e2:sf3:xxxe")},
	{Entity: TestS1{"xxx", 2, "66", 0}, Data: []byte("d4:XXYhi556e4:ffgdi2e3:hhd2:662:sf3:xxxe")},
	{Entity: TestS2{GGG: "gggee", IFace: []interface{}{4, 6}}, Data: []byte("d3:GGG5:gggee5:ifaceli4ei6eee")},
	{Entity: TestS2{GGG: "gggee", IFace: map[string]interface{}{"gds": 8, "353": -45}}, Data: []byte("d3:GGG5:gggee5:ifaced3:353i-45e3:gdsi8eee")},
	{Entity: []int{1, 3, 5}, Data: []byte("li1ei3ei5ee")},
}

func TestMarshal(t *testing.T) {
	tests := append(comTests, encTests...)
	for i, tt := range tests {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			got, err := Marshal(tt.Entity)
			if err != nil || !reflect.DeepEqual(got, tt.Data) {
				t.Errorf("Marshal() = %s, Data = %s, err = %v", got, tt.Data, err)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := append(comTests, decTests...)
	for i, tt := range tests {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			rv := reflect.New(reflect.TypeOf(tt.Entity))
			err := Unmarshal(tt.Data, rv.Interface())
			got := rv.Elem().Interface()
			if err != nil || !reflect.DeepEqual(got, tt.Entity) {
				t.Errorf("Unmarshal() = %v, Entity = %v, err = %v", got, tt.Entity, err)
			}
		})
	}
}
