package db

import (
	"DFS/db/mysql"
	"DFS/model"
	"log"
	"time"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/25 8:14 下午
 * @Desc: 用户文件表进行操作
 */

func UpdateUserFile(username string, fileMetaData model.FileMetaData) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("insert ignore into tbl_user_file(user_name, file_sha1, file_size, file_name, upload_at, status) values (?, ?, ?, ?, ?, 1)")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return false
	}
	r, err := stmt.Exec(username, fileMetaData.FileHash, fileMetaData.FileSize, fileMetaData.FileName, time.Now())
	if _, err := r.RowsAffected(); nil == err {
		log.Println("----------------------------用户文件表插入成功----------------------------")
		return true
	}
	log.Println("----------------------------用户文件表插入失败----------------------------", err)
	return false
}
