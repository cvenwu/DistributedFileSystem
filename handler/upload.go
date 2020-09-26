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
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/26 9:31 上午
 * @Desc: 普通上传文件
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
		//TODO:这里是否需要向文件表添加记录呢？目前我们这里是基于文件表中没有记录，因为一旦有记录就可以直接秒传来做
		opRet := dblayer.AddFileMetaData(fileMetaData)
		r.ParseForm()
		username := r.Form.Get("username")
		upRet := dblayer.UpdateUserFile(username, fileMetaData)
		if !opRet && !upRet {
			log.Println("--------------------------文件表或用户文件表插入记录失败----------------------------")
			io.WriteString(w, "文件表或用户文件表插入记录失败")
		} else {
			//4. 返回响应表示我们上传成功
			//重定向：返回一个302响应码，要求用户向响应头部中的url重新发起请求
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
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
	log.Println("用户点击了下载文件，文件hash码为：", filehash)
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
	log.Println("下载文件成功--------------------------------")
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
	username := r.Form.Get("username")
	ret, err := dblayer.GetLastestFileMetaData(username, limitCount)
	if err != nil {
		log.Println("----------------------------获取用户最近上传文件失败----------------------------")
		w.Write([]byte("获取用户最近上传文件失败，请稍后再试！！！"))
		return
	}
	log.Println(ret)
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

/*
实现秒传：需要用户传入一个特征值，如果数据库中有这个特征值，就实现秒传
*/
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析请求参数
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	username := r.Form.Get("username")

	//2. 判断一下能否触发秒传，也就是传递的filehash是否已经存在于文件表中
	//TODO ：这里其实应该将文件的filehash存在于redis缓存中
	fileMetaData, isFast := IsFastUpload(filehash)
	if !isFast {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	//4. 如果上传过该文件，只写入记录到文件信息表，并且返回秒传成功
	opType := dblayer.UpdateUserFile(username, fileMetaData)
	if !opType {
		log.Println("----------------------------秒传成功但用户文件表插入失败，请稍后再试----------------------------")
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传成功但用户文件表插入失败，请稍后再试",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "秒传成功",
	}
	w.Write(resp.JSONBytes())
	return
}

/*
判断能否触发秒传
返回两个值：第1个是如果可以秒传，从数据库中查询到的对应文件的信息
			第2个是判断是否可以秒传
*/
func IsFastUpload(filehash string) (model.FileMetaData, bool) {
	//从文件表查询是否有相同hash值的文件记录
	fileMetaData, err := dblayer.GetFileMetaData(filehash)
	if err != nil {
		log.Println("----------------------------根据用户传入的filehash查找对应的文件失败，请稍后再试----------------------------")
		return model.FileMetaData{}, false
	}

	//如果查不到记录则秒传失败
	if fileMetaData.FileName == "" {
		return model.FileMetaData{}, false
	}
	log.Println("----------------------------秒传成功----------------------------")
	return fileMetaData, true
}

/*
普通上传文件
*/
func UploadFile(f multipart.File, filename string, username string) bool {
	//本地新建一个文件用来存储用户上传的文件
	localFile, err := os.Create("./tmp/" + filename)
	defer localFile.Close() //记得操作完成之后关闭文件
	if err != nil {
		log.Println("--------------------------云端创建要存储的文件失败----------------------------")
		return false
	}
	writeSize, err := io.Copy(localFile, f)

	if err != nil {
		log.Println("--------------------------云端创建写入文件失败----------------------------")
		return false
	}
	//将文件写入到我们的存储地址，同时向数据库表中添加记录
	//新建一个文件结构体
	fileMetaData := model.FileMetaData{
		FileName:     filename,
		UploadTimeAt: time.Now().Format("2006-01-02 15:04:05"),
		FileLocation: "./tmp/" + filename, //TODO:这里如果这样写代码，如何确保用户不会上传同一个文件名的不同文件造成文件覆盖
	}
	fileMetaData.FileSize = writeSize
	localFile.Seek(0, 0)
	//因为要计算文件哈希值需要从文件头部开始计算，所以要文件的游标移动到文件的最开始位置处
	fileMetaData.FileHash = util.FileSha1(localFile)
	log.Println("上传文件的sha1值为-------------------->", fileMetaData.FileHash)

	//向数据库添加一条文件记录
	//TODO:这里是否需要向文件表添加记录呢？目前我们这里是基于文件表中没有记录，因为一旦有记录就可以直接秒传来做
	opRet := dblayer.AddFileMetaData(fileMetaData)
	upRet := dblayer.UpdateUserFile(username, fileMetaData)
	if !opRet || !upRet {
		log.Println("--------------------------文件表或用户文件表插入记录失败----------------------------")
		return false
	}
	return true
}
