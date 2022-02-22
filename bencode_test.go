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

var comTests = []TestCase{
	{Entity: 23, Data: []byte("i23e")},
	{Entity: -422, Data: []byte("i-422e")},
	{Entity: -0, Data: []byte("i0e")},
	{Entity: "abc", Data: []byte("3:abc")},
	{Entity: "", Data: []byte("0:")},
	{Entity: "3242434te", Data: []byte("9:3242434te")},
	{Entity: []int{1, 2, 3, 4, 5}, Data: []byte("li1ei2ei3ei4ei5ee")},
	{Entity: []string{"aa", "b", "ccc"}, Data: []byte("l2:aa1:b3:ccce")},
	{Entity: []interface{}{"aa", "b", 33, -23, "XX"}, Data: []byte("l2:aa1:bi33ei-23e2:XXe")},
}

var encTests = []TestCase{
	{Entity: TestS1{"xxx", 2, "", 556}, Data: []byte("d2:sf3:xxx4:ffgdi2ee")},
	{Entity: TestS1{"xxx", 2, "66", 556}, Data: []byte("d2:sf3:xxx4:ffgdi2e3:hhd2:66e")},
}

var decTests = []TestCase{
	{Entity: TestS1{"xxx", 2, "", 0}, Data: []byte("d2:sf3:xxx4:ffgdi2ee")},
	{Entity: TestS1{"xxx", 2, "66", 0}, Data: []byte("d2:sf3:xxx4:ffgdi2e3:hhd2:664:XXYhi556ee")},
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
