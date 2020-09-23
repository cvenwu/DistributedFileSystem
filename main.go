package main

import (
	"DFS/handler"
	"net/http"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/22 8:47 下午
 * @Desc:
 */



/*
1. 上传文件
2. 下载文件
 */

func main() {
	//上传文件
	http.HandleFunc("/file/upload", handler.UploadFileHandler)
	//上传文件成功对应的监听函数
	http.HandleFunc("/file/upload/suc", handler.UploadFileSucHandler)
	//下载文件
	http.HandleFunc("/file/download", handler.DownloadFileHandler)
	//查询文件元信息
	http.HandleFunc("/file/meta", handler.GetFileMetaData)

	//删除文件元信息
	http.HandleFunc("/file/delete", handler.DeleteFileMetaData)

	//修改文件元信息【文件重命名】
	http.HandleFunc("/file/update", handler.UpdateFileMetaData)
	//获取用户最近上传文件【批量查询】
	http.HandleFunc("/file/query", handler.GetLatestFileMetaData)



	//监听服务器端口
	http.ListenAndServe(":8080", nil)
}