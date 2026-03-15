package gtmysql

import (
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
	"gorm.io/gorm"
)

// ExplainSQL 生成完整的 SQL 语句，包含参数值，便于调试和日志记录
func ExplainSQL(tx *gorm.DB) string {
	sql := tx.Statement.SQL.String()
	vars := tx.Statement.Vars
	return tx.Dialector.Explain(sql, vars...)
}

// GetCustomizedConfigStrFields 获取自定义配置的字段字符串，适用于 SQL 建表语句中的字段列表
func GetCustomizedConfigStrFields(c *gtconfpb.MysqlTableFieldCustomizedConfig) string {
	if c == nil {
		return ""
	}
	res := ""
	for _, s := range c.Fields {
		res += "\n    " + s + ","
	}
	return res
}

// GetCustomizedConfigStrIndexes 获取自定义配置的索引字符串，适用于 SQL 建表语句中的索引列表
func GetCustomizedConfigStrIndexes(c *gtconfpb.MysqlTableFieldCustomizedConfig) string {
	if c == nil {
		return ""
	}
	res := ""
	for _, s := range c.Indexes {
		res += ",\n    " + s
	}
	return res
}
