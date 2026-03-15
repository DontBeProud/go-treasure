package gtmysql

import "testing"

func TestMap2MysqlJsonString(t *testing.T) {
	if res := Map2MysqlJsonString(map[string]interface{}{}); res != "{}" {
		panic(res)
	}
	if res := Map2MysqlJsonString(map[string]interface{}{"key": "value"}); res == "{}" {
		panic(res)
	}
}
