package gtmysql

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// SetWhereOrCondition 设置 where 内部多个 or 的查询条件
func SetWhereOrCondition(db *gorm.DB, rawFmtList []string, fmtParams ...interface{}) *gorm.DB {
	condition := ""
	for _, rawFmt := range rawFmtList {
		tmp := strings.TrimSpace(rawFmt)
		if tmp == "" {
			continue
		}
		condition += fmt.Sprintf(" or (%s)", tmp)
	}
	if len(condition) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("(%s)", strings.TrimPrefix(condition, " or ")), fmtParams...)
}

// SetWhereInCondition 设置 where in 查询条件
func SetWhereInCondition(db *gorm.DB, fieldName string, v interface{}) *gorm.DB {
	return db.Where(fmt.Sprintf("%s in ?", fieldName), v)
}

// SetWhereEqualCondition 设置 where = 查询条件，当 v 为 nil 指针时忽略该条件
func SetWhereEqualCondition(db *gorm.DB, fieldName string, v interface{}) *gorm.DB {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr && isNil(v) {
		return db
	}
	return db.Where(fmt.Sprintf("%s = ?", fieldName), v)
}

// SetWhereRegExpCondition 设置 where regexp 查询条件
func SetWhereRegExpCondition(db *gorm.DB, fieldName string, reg string, ignoreEmptyString bool) *gorm.DB {
	if ignoreEmptyString && reg == "" {
		return db
	}
	return db.Where(fmt.Sprintf("%s regexp ?", fieldName), reg)
}

func isNil(i interface{}) bool {
	defer func() { _ = recover() }()
	return reflect.ValueOf(i).IsNil()
}
