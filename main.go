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

	//监听服务器端口
	http.ListenAndServe(":8080", nil)
}