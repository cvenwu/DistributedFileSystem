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
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assets.AssetFS())))
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	//上传文件
	http.HandleFunc("/file/upload", handler.HttpInterceptor(handler.UploadFileHandler))
	//上传文件成功对应的监听函数
	http.HandleFunc("/file/upload/suc", handler.HttpInterceptor(handler.UploadFileSucHandler))
	//下载文件
	http.HandleFunc("/file/downloadurl", handler.HttpInterceptor(handler.DownloadFileHandler))
	//查询文件元信息
	http.HandleFunc("/file/meta", handler.HttpInterceptor(handler.GetFileMetaData))
	//删除文件元信息
	http.HandleFunc("/file/delete", handler.HttpInterceptor(handler.DeleteFileMetaData))
	//修改文件元信息【文件重命名】
	http.HandleFunc("/file/update", handler.HttpInterceptor(handler.UpdateFileMetaData))
	//获取用户最近上传文件【批量查询】
	http.HandleFunc("/file/query", handler.HttpInterceptor(handler.GetLatestFileMetaData))

	//--------------处理用户相关操作
	//用户注册
	http.HandleFunc("/user/signup", handler.UserSignUp)
	//用户登录
	http.HandleFunc("/user/signin", handler.UserSignIn)
	//查询用户信息
	http.HandleFunc("/user/info", handler.HttpInterceptor(handler.UserInfoHandler))

	//监听服务器端口
	http.ListenAndServe(":8080", nil)
}
