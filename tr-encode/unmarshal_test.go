package trencode

import (
	"testing"
	"time"
)

func TestUnmarshal2Map(t *testing.T) {
	data := []byte(`{"a":1,"b":2}`)
	m, err := Unmarshal2Map("json", data)
	if err != nil {
		panic(err)
	}
	for k, v := range m {
		println(k, v)
	}

	b, _ := time.ParseInLocation("20060102", "20250610", time.Local)
	println(time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.Local).Add(-1).Unix())
}
