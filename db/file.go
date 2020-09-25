package db

import (
	"DFS/db/mysql"
	"DFS/model"
	"DFS/util"
	"log"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 10:07 上午
 * @Desc: 对数据库中文件表进行操作
 */

/*
根据文件的filehash值获取对应的文件结构体
*/
func GetFileMetaData(filehash string) (model.FileMetaData, error) {
	stmt, err := mysql.GetMysqlConn().Prepare("Select file_sha1, file_name, file_size, file_addr, DATE_FORMAT(update_at,'%Y-%m-%d %H:%i:%S') from tbl_file where file_sha1 = ? and status = 1 limit 1")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return model.FileMetaData{}, err
	}
	//定义一个
	fileMetaData := model.FileMetaData{}
	err = stmt.QueryRow(filehash).Scan(&fileMetaData.FileHash, &fileMetaData.FileName, &fileMetaData.FileSize, &fileMetaData.FileLocation, &fileMetaData.UploadTimeAt)
	fileMetaData.FileSizeFormat = util.FormatFileSize(fileMetaData.FileSize)
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

//删除文件
func DeleteFileMetaData(filehash string) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("update tbl_file set status = 0 where file_sha1 = ?")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return false
	}

	ret, err := stmt.Exec(filehash)
	if err != nil {
		log.Println("----------------------------执行删除语句失败----------------------------")
		return false
	}

	rowsAffectNum, err := ret.RowsAffected()
	if rowsAffectNum <= 0 {
		return false
	}

	return true
}

//修改文件信息
func UpdateFileMetaData(filehash string, filename string) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("update tbl_file set file_name = ? where file_sha1 = ? and status = 1")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return false
	}
	ret, err := stmt.Exec(filename, filehash)
	if err != nil {
		log.Println("----------------------------文件表修改记录失败----------------------------")
		return false
	}
	//判断是否修改成功
	if num, err := ret.RowsAffected(); nil == err {
		if num > 0 {
			log.Println("----------------------------文件表修改记录成功----------------------------")
			return true
		}
	}
	//走到这里说明有错误，没有插入成功
	return false
}

/*
查询用户最近上传的limit条文件记录
 */
func GetLastestFileMetaData(username string, limit int) ([]model.FileMetaData, error) {
	stmt, err := mysql.GetMysqlConn().Prepare("Select file_sha1, file_name, file_size, DATE_FORMAT(upload_at,'%Y-%m-%d %H:%i:%S'), DATE_FORMAT(last_update,'%Y-%m-%d %H:%i:%S') from tbl_user_file where user_name = ? and status = 1 order by last_update desc limit ?")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------", err)
		return []model.FileMetaData{}, err
	}

	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println("----------------------------查询后赋值失败----------------------------", err)
		return []model.FileMetaData{}, err
	}

	ret := []model.FileMetaData{}
	for rows.Next() {
		//定义一个
		fileMetaData := model.FileMetaData{}
		rows.Scan(&fileMetaData.FileHash, &fileMetaData.FileName, &fileMetaData.FileSize, &fileMetaData.UploadTimeAt, &fileMetaData.LastUpdated)
		fileMetaData.FileSizeFormat = util.FormatFileSize(fileMetaData.FileSize)
		ret = append(ret, fileMetaData)
	}

	return ret, nil
}
