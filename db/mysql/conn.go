package mysql

import (
	config "DFS/config"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 9:37 上午
 * @Desc:
 */

var conn *sql.DB

//初始化连接
func init() {
	dataBaseSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&loc=Local&parseTime=true", config.MysqlUser,
		config.MysqlPassword, config.MysqlHost, config.MysqlPort, config.MysqlDatabase, config.MysqlCharset)
	conn, _ = sql.Open("mysql", dataBaseSource)

	//测试连接连接是否成功
	if conn.Ping() != nil {
		log.Println("----------------------------获取MySql连接失败----------------------------")
		//进程非正常退出
		os.Exit(1)
	}

	log.Println("----------------------------MySql连接成功----------------------------")

}

//返回一个连接对象
func GetMysqlConn() *sql.DB {
	if conn != nil {
		log.Println("----------------------------获取MySql连接成功----------------------------")
		return conn
	}

	log.Println("----------------------------获取MySql连接失败----------------------------")
	return nil
}
