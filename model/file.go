package model


/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 9:48 上午
 * @Desc: 文件信息描述
 */

type FileMetaData struct {
	FileHash     string //文件哈希值
	FileName     string //文件名
	FileSize     int64  //文件大小
	FileLocation string //文件存储目录
	UploadTimeAt string //上传时间
}
