package handler

import (
	dblayer "DFS/db"
	"DFS/model"
	"DFS/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 9:35 上午
 * @Desc:
 */

/*
上传文件
*/
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	//如果用户是get请求，就直接读取上传页面并返回
	if r.Method == http.MethodGet {
		content, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			log.Println("--------------------------打开index.html失败----------------------------")
			io.WriteString(w, "Internal Server Error")
			return
		}
		io.WriteString(w, string(content))
	} else if r.Method == http.MethodPost { //如果是post请求，说明用户要将文件传送到云端存储
		//1. 解析用户请求表单，获取用户携带的文件
		//返回3个参数，第1个参数表示文件句柄，第2个参数表示用户上传的文件的文件头部
		f, fHeader, err := r.FormFile("file")
		if err != nil {
			log.Println("--------------------------获取用户上传文件信息失败----------------------------")
			w.Write([]byte("获取用户上传文件信息失败"))
		}

		//3. 将文件写入到我们的存储地址，同时向数据库表中添加记录
		//新建一个文件结构体
		fileMetaData := model.FileMetaData{
			FileName:     fHeader.Filename,
			UploadTimeAt: time.Now().Format("2006-01-02 15:04:05"),
			FileLocation: "./tmp/" + fHeader.Filename, //TODO:这里如果这样写代码，如何确保用户不会上传同一个文件名的不同文件造成文件覆盖
		}
		//本地新建一个文件用来存储用户上传的文件
		localFile, err := os.Create("./tmp/" + fHeader.Filename)
		defer localFile.Close() //记得操作完成之后关闭文件
		if err != nil {
			log.Println("--------------------------云端创建要存储的文件失败----------------------------")
			io.WriteString(w, "云端创建要存储的文件失败")
		}
		writeSize, err := io.Copy(localFile, f)
		if err != nil {
			log.Println("--------------------------云端创建写入文件失败----------------------------")
			io.WriteString(w, "云端创建写入文件失败")
		}
		fileMetaData.FileSize = writeSize
		localFile.Seek(0, 0)
		//因为要计算文件哈希值需要从文件头部开始计算，所以要文件的游标移动到文件的最开始位置处
		fileMetaData.FileHash = util.FileSha1(localFile)
		log.Println("上传文件的sha1值为-------------------->", fileMetaData.FileHash)

		//向数据库添加一条文件记录

		opRet := dblayer.AddFileMetaData(fileMetaData)
		if !opRet {
			log.Println("--------------------------文件表插入记录失败----------------------------")
			io.WriteString(w, "文件表插入记录失败")
		} else {
			//4. 返回响应表示我们上传成功
			//重定向：返回一个302响应码，要求用户向响应头部中的url重新发起请求
			http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
		}
	}
}

/*
下载文件：用户需要传入一个filehash值
1. 解析用户请求并获取对应的filehash
2. 查询对应filehash是否存在，如果不存在，就返回不存在
3.						  如果存在，就得到对应文件的信息
									打开对应的文件，然后读取文件内容并返回，
*/
func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析用户请求并获取对应的filehash
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	//2. 根据filehash获取对应的文件元信息
	fileMetaData, err := dblayer.GetFileMetaData(filehash)
	if err != nil {
		log.Println("--------------------------文件元信息获取失败----------------------------")
		w.Write([]byte("文件元信息获取失败，请稍后再试"))
		return
	}
	log.Println(fileMetaData)
	//3. 如果存在，就根据文件路径打开文件
	f, err := os.Open(fileMetaData.FileLocation)
	if err != nil {
		log.Println("--------------------------打开文件失败----------------------------")
		w.Write([]byte("文件不存在，请稍后再试"))
		return
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("--------------------------读取文件内容失败，请稍后再试----------------------------")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//理论上工作是已经做完了的，但是为了让浏览器做一个演示，我们需要将一个http的响应头，让浏览器识别出来就可以当成一个文件进行下载
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMetaData.FileName))

	w.Write(data)
}

/*
上传成功
*/
func UploadFileSucHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("文件上传成功............."))
}

/*
查询文件元信息
*/
func GetFileMetaData(w http.ResponseWriter, r *http.Request) {
	//1. 获取用户表单filesha1值
	r.ParseForm()
	fileHash := r.Form.Get("filehash")
	if len(fileHash) <= 0 {
		w.Write([]byte("请输入文件的哈希值："))
		return
	}
	//2. 根据获取用户filesha1值查询对应的文件元信息
	fileMetaData, err := dblayer.GetFileMetaData(fileHash)
	if err != nil {
		log.Println("----------------------------根据filehash获取文件元信息失败----------------------------")
		w.Write([]byte("根据filehash获取文件元信息失败"))
		return
	}
	//3. 将文件元信息序列化
	ret, err := json.Marshal(fileMetaData)
	if err != nil {
		log.Println("----------------------------文件结构信息序列化失败----------------------------")
		w.Write([]byte("文件结构信息序列化失败"))
		return
	}
	//4. 将序列化后的信息返回给用户
	w.Write(ret)
}

/*
查询最近上传的几个文件的信息
*/
func GetLatestFileMetaData(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCount, err := strconv.Atoi(r.Form.Get("limit"))
	if err != nil {
		log.Println("----------------------------获取用户获取的文件个数失败----------------------------")
		w.Write([]byte("获取用户获取的文件个数失败，请稍后再试！！！"))
		return
	}
	ret, err := dblayer.GetLastestFileMetaData(limitCount)
	if err != nil {
		log.Println("----------------------------获取用户最近上传文件失败----------------------------")
		w.Write([]byte("获取用户最近上传文件失败，请稍后再试！！！"))
		return
	}
	wRet, err := json.Marshal(ret)
	if err != nil {
		log.Println("----------------------------json序列化失败----------------------------")
		w.Write([]byte("json序列化失败，请稍后再试！！！"))
		return
	}
	w.Write(wRet)
}

/*
文件修改：post接口，目前只是文件重命名
*/
func UpdateFileMetaData(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//1. 解析用户请求：获取要修改文件的hash值以及文件新的名字
		r.ParseForm()
		filehash := r.Form.Get("filehash")

		filename := r.Form.Get("filename")
		opRet := dblayer.UpdateFileMetaData(filehash, filename)
		if !opRet {
			log.Println("----------------------------文件重命名失败----------------------------")
			w.Write([]byte("文件重命名失败，请稍后再试！！！"))
			return
		}
		w.Write([]byte("文件重命名成功！！！"))
	}
}

/*
文件删除：目前是get请求
用户需要传入一个filehash参数
*/
func DeleteFileMetaData(w http.ResponseWriter, r *http.Request) {
	//1. 获取用户表单filesha1值
	r.ParseForm()
	fileHash := r.Form.Get("filehash")
	if len(fileHash) <= 0 {
		w.Write([]byte("请输入文件的哈希值："))
		return
	}
	//2. 根据获取用户filesha1值查询对应的文件元信息
	opRet := dblayer.DeleteFileMetaData(fileHash)
	if !opRet {
		log.Println("----------------------------文件删除失败----------------------------")
		w.Write([]byte("文件删除失败，请稍后再试"))
		return
	}
	w.Write([]byte("文件删除成功"))
}
