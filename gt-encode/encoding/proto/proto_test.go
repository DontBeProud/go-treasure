package proto

import (
	"reflect"
	"testing"
)

func TestName(t *testing.T) {
	c := new(Codec)
	if !reflect.DeepEqual(c.Name(), "proto") {
		t.Errorf("no expect float_key value: %v, but got: %v", c.Name(), "proto")
	}
}
