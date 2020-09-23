package db

import (
	"DFS/db/mysql"
	"DFS/model"
	"log"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 10:07 上午
 * @Desc: 对数据库中文件表进行操作
 */

func GetFileMetaData(filehash string) (model.FileMetaData, error) {
	stmt, err := mysql.GetMysqlConn().Prepare("Select file_sha1, file_name, file_size, file_addr, update_at from tbl_file where file_sha1 = ? and status = 1 limit 1")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return model.FileMetaData{}, err
	}
	//定义一个
	fileMetaData := model.FileMetaData{}
	err = stmt.QueryRow(filehash).Scan(&fileMetaData.FileHash, &fileMetaData.FileName, &fileMetaData.FileSize, &fileMetaData.FileLocation, &fileMetaData.UploadTimeAt)

	if err != nil {
		log.Println("----------------------------查询后赋值失败----------------------------", err)
		return model.FileMetaData{}, err
	}

	return fileMetaData, nil
}

//向文件表插入一条记录：返回true或false表明操作是否成功
func AddFileMetaData(fileMetaData model.FileMetaData) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("INSERT INTO tbl_file(file_sha1, file_name, file_size, file_addr, create_at, update_at, status) values(?, ?, ?, ?, ?, ?, 1)")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return false
	}
	ret, err := stmt.Exec(fileMetaData.FileHash, fileMetaData.FileName, fileMetaData.FileSize, fileMetaData.FileLocation, fileMetaData.UploadTimeAt, fileMetaData.UploadTimeAt)
	if err != nil {
		log.Println("----------------------------文件表插入记录失败----------------------------")
		return false
	}
	//判断数据是否重新插入
	if num, err := ret.RowsAffected(); nil == err {
		if num <= 0 {
			log.Println("----------------------------数据之前已经插入：----------------------------")
		}
		return true
	}
	//走到这里说明有错误，没有插入成功
	return false
}
