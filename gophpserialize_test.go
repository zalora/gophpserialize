package gophpserialize

import "testing"

func TestUnmarshal(t *testing.T) {
	data := `a:3:{s:5:"apple";i:1;s:6:"orange";i:2;s:5:"grape";i:3;}`

	obj := Unmarshal([]byte(data))

	if obj["apple"] != 1 {
		t.Error("Unmarshal failed")
	}
	if obj["orange"] != 2 {
		t.Error("Unmarshal failed")
	}
	if obj["grape"] != 3 {
		t.Error("Unmarshal failed")
	}
}
