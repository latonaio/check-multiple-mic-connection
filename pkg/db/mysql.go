package db

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"log"

	"check-multiple-mic-connection/config"
	_ "gorm.io/driver/mysql"
)

var mysql *MysqlDB

type MysqlDB struct {
	conn *gorm.DB
}

func (md *MysqlDB) Connect(config *config.Config) DB {
	dbConnectInfo := fmt.Sprintf(
		`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local`,
		config.DB.DbUserName,
		config.DB.DbUserPassword,
		config.DB.DbHost,
		config.DB.DbPort,
		config.DB.DbName,
	)
	sqlDB, err := gorm.Open("mysql", dbConnectInfo)
	if err != nil {
		log.Printf("db connection error. %s", err)
	}

	mysql = &MysqlDB{conn: sqlDB}

	return mysql
}

func (md *MysqlDB) GetConn() *gorm.DB {
	return md.conn
}

func (md *MysqlDB) Close() error {
	return md.conn.Close()
}

func GetMysql() *MysqlDB{
	return mysql
}

