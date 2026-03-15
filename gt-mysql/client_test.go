package gtmysql

import (
	"os"
	"testing"

	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
)

func skipTest(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_TEST") != "true" {
		t.Skip("skipping test; set RUN_TEST=true to run")
	}
}

func TestNewClient(t *testing.T) {
	skipTest(t)

	client, _, err := NewClient(&gtconfpb.MysqlConfig{
		Con: &gtconfpb.MysqlConnectConfig{
			Ip:       "127.0.0.1",
			Port:     3306,
			UserName: "root",
			Password: "your_password",
			DbName:   "gt_test",
		},
	}, nil, nil)
	if err != nil {
		panic(err.Error())
	}

	if err = client.Exec("CREATE TABLE IF NOT EXISTS `table_gt_test` ( `id` INT(10) NOT NULL AUTO_INCREMENT, " +
		"`name` VARCHAR(50) NOT NULL DEFAULT '' COLLATE 'utf8mb4_general_ci', " +
		"PRIMARY KEY (`id`) USING BTREE, INDEX `name` (`name`) USING BTREE) COLLATE='utf8mb4_general_ci' " +
		"ENGINE=InnoDB ;").Error; err != nil {
		panic(err.Error())
	}

	if err = client.Table("table_gt_test").Create(&struct{ Name string }{Name: "test"}).Error; err != nil {
		panic(err.Error())
	}

	var id *int
	if err = client.Table("table_gt_test").Select("id").
		Where("name = ?", "test_not_exist").Limit(1).Scan(id).Error; err != nil {
		panic(err.Error())
	}
	println(id == nil)
}
