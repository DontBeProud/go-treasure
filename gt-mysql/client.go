package gtmysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	gtlog "github.com/DontBeProud/go-treasure/gt-log"
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewClient 创建mysql-gorm对象
func NewClient(config *gtconfpb.MysqlConfig, gormConfig *gorm.Config, logger gtlog.Logger) (*gorm.DB, func(), error) {
	conn, cleanup, err := NewConn(config, logger)
	if err != nil {
		return nil, func() {}, err
	}
	if gormConfig == nil {
		gormConfig = &gorm.Config{}
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: conn,
	}), gormConfig)
	if err != nil {
		return nil, func() {}, err
	}
	return db, cleanup, err
}

// ForkDB fork一个新的mysql-gorm对象
// 使用 db.Session(&gorm.Session{}) 创建独立会话，与原对象完全隔离，避免并发污染。
func ForkDB(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{})
}

// ForkDBWithContext fork一个新的mysql-gorm对象，并设置ctx
func ForkDBWithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx)
}

// NewConn 创建mysql数据库连接（*sql.DB）
func NewConn(config *gtconfpb.MysqlConfig, logger gtlog.Logger) (db *sql.DB, cleanup func(), err error) {
	if config == nil {
		return nil, func() {}, errors.New("mysql config is nil")
	}
	if config.Con == nil {
		return nil, func() {}, errors.New("mysql config con is nil")
	}

	dsn, err := parseDSN(config.Con)
	if err != nil {
		return nil, func() {}, err
	}

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, func() {}, err
	}

	if config.MaxOpenCon != nil {
		sqlDB.SetMaxOpenConns(int(*config.MaxOpenCon))
	}
	if config.MaxIdleCon != nil {
		sqlDB.SetMaxIdleConns(int(*config.MaxIdleCon))
	}
	if config.MaxIdleTime != nil {
		sqlDB.SetConnMaxIdleTime(config.MaxIdleTime.AsDuration())
	}
	if config.MaxLifeTime != nil {
		sqlDB.SetConnMaxLifetime(config.MaxLifeTime.AsDuration())
	}

	pingTimeout := defaultPingTimeout
	if config.PingTimeout != nil {
		if d := config.PingTimeout.AsDuration(); d > 0 {
			pingTimeout = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err = sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, func() {}, fmt.Errorf("mysql ping failed: %w", err)
	}

	cleanup = func() { _ = sqlDB.Close() }
	return sqlDB, cleanup, nil
}

// parseDSN 根据配置解析DSN
func parseDSN(config *gtconfpb.MysqlConnectConfig) (string, error) {
	if config == nil {
		return "", errors.New("parseDSN err: connect config is nil")
	}

	// 若有配置dsn，则直接返回
	if config.Dsn != "" {
		return config.Dsn, nil
	}

	if config.Ip == "" {
		return "", errors.New("parseDSN err: IP未配置")
	}
	if config.Port == 0 {
		return "", errors.New("parseDSN err: 端口未配置")
	}
	if config.DbName == "" {
		return "", errors.New("parseDSN err: 数据库名未配置")
	}

	netWork := "tcp"
	if n := config.GetNetwork(); n != "" {
		netWork = n
	}

	parseTime := !config.DontParseTime

	charset := "utf8mb4"
	if c := config.GetCharset(); c != "" {
		charset = c
	}

	loc := make([]string, 0)
	if l := config.GetLocation(); l != "" {
		decoded, err := url.PathUnescape(l)
		if err != nil {
			return "", errors.New("parseDSN err: location配置解析异常: " + err.Error())
		}
		loc = append(loc, decoded)
	}

	options := &url.Values{
		"charset":   []string{charset},
		"parseTime": []string{fmt.Sprintf("%v", parseTime)},
		"loc":       loc,
	}
	for k, v := range config.Options {
		options.Set(k, v)
	}

	// "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	return fmt.Sprintf("%s:%s@%s(%s:%d)/%s?%s",
		config.UserName, config.Password, netWork, config.Ip, config.Port, config.DbName, options.Encode()), nil
}
