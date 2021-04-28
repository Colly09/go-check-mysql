package dbs

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dbService *gorm.DB

type Result struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default string
	Extra   string
}

func GetCreate(tableName string) string {
	// var tableInfo []Create
	rows, err := dbService.Raw(fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)).Rows()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer rows.Close()

	table := ""
	info := ""
	for rows.Next() {
		rows.Scan(&table, &info)
	}
	return info
}

func GetInfo(tableName string) []Result {
	var tableInfo []Result
	// dbService.Raw(fmt.Sprintf("DESCRIBE `%s`", tableName)).Scan(&tableInfo)
	rows, err := dbService.Raw(fmt.Sprintf("DESCRIBE `%s`", tableName)).Rows()
	if err != nil {
		fmt.Println(err)
		return make([]Result, 0)
	}
	defer rows.Close()
	for rows.Next() {
		dbService.ScanRows(rows, &tableInfo)
	}
	// fmt.Println(tableInfo)
	return tableInfo
}

func ConnMysql(password string, dbName string) {
	db, err := gorm.Open(mysql.Open(fmt.Sprintf("%s/%s?charset=utf8&parseTime=True&loc=Local", password, dbName)), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
	dbService = db
}
