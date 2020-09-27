package handler

import (
	rPool "DFS/cache/redis"
	dblayer "DFS/db"
	"DFS/model"
	"DFS/util"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
	"strings"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/27 8:50 上午
 * @Desc: 秒传相关的业务逻辑
 */

/*
实现秒传：需要用户传入一个特征值，如果数据库中有这个特征值，就实现秒传
*/
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析请求参数
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	username := r.Form.Get("username")

	//2. 判断一下能否触发秒传，也就是传递的filehash是否已经存在于文件表中
	//（不需要这样，因为种类一旦秒传成功说明就在redis缓存中，如果秒传不成功就会让用户访问普通上传接口）TODO ：这里其实应该将文件的filehash存在于redis缓存中
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
返回两个值：第1个是如果可以秒传，从redis中查询到的文件元信息，只有一个文件路径以及文件名，文件哈希
			第2个是判断是否可以秒传
*/
func IsFastUpload(filehash string) (model.FileMetaData, bool) {
	//从文件表查询是否有相同hash值的文件记录
	rConn := rPool.GetRedisConn().Get()
	data, err := redis.String(rConn.Do("hget", "tbl_file",  "tbl_file:" + filehash))

	fileLocation := strings.Split(data, "/")
	fileName := fileLocation[len(fileLocation)-1]
	fileMetaData := model.FileMetaData{
		FileHash: filehash,
		FileName: fileName,
		FileLocation: data,
	}

	if err != nil {
		log.Println("----------------------------根据用户传入的filehash查找对应的文件失败，请稍后再试----------------------------")
		return model.FileMetaData{}, false
	}

	//如果查不到记录则秒传失败
	if fileMetaData.FileName == "" {
		return model.FileMetaData{}, false
	}
	log.Println("----------------------------可以秒传----------------------------")
	return fileMetaData, true
}
