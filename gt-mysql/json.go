package gtmysql

import "github.com/bytedance/sonic"

// Map2MysqlJsonString map转换成合法的mysql json类型数据
// nil/empty/error => "{}"
func Map2MysqlJsonString(m map[string]interface{}) string {
	if m != nil {
		if j, err := sonic.Marshal(m); err == nil {
			return string(j)
		}
	}
	return "{}"
}
