package handler

import (
	rPool "DFS/cache/redis"
	"DFS/config"
	dblayer "DFS/db"
	"DFS/model"
	"DFS/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 9:35 上午
 * @Desc: 分块上传相关
 */

//初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadId   string //上传的id号
	ChunkSize  int    //分块的大小
	ChunkCount int
}

/*
初始化分块信息
*/
func InitializeMulpartInfo(filehash string, filesize int, username string, multiUploadInfo *MultipartUploadInfo) bool {

	//2. 获得redis的一个连接通过Get
	rConn := rPool.GetRedisConn().Get()
	//不要忘记在退出的时候关闭redis的连接
	defer rConn.Close()

	//3. 生成分块上传的初始化信息，这里我们将初始化信息封装成一个结构体
	multiUploadInfo.FileSize = filesize
	multiUploadInfo.FileHash = filehash
	multiUploadInfo.UploadId = username + fmt.Sprintf("%x", time.Now().UnixNano()) //我们简单地定义一个规则：当前用户名加上当前时间戳
	multiUploadInfo.ChunkSize = config.ChunkSize                                   //每一块的大小，我们自己在配置文件中已经定义好了
	multiUploadInfo.ChunkCount = int(math.Ceil(float64(filesize) / config.ChunkSize))

	log.Println("----------------------------生成分块上传的初始化信息成功----------------------------")

	//4. 将初始化的信息写入到redis的缓存中去，通过hset命令来将数据缓存起来
	//采用HMSET一次性将这些指令全部写进去，
	rConn.Do("HMSET", "MP_"+multiUploadInfo.UploadId, "chunkcount", multiUploadInfo.ChunkCount, "filehash",
		multiUploadInfo.FileHash, "filesize", multiUploadInfo.FileSize)

	//rConn.Do("HSET", "MP_"+multiUploadInfo.UploadId, "chunkcount", multiUploadInfo.ChunkCount)
	//rConn.Do("HSET", "MP_"+multiUploadInfo.UploadId, "filehash", multiUploadInfo.FileHash)
	//rConn.Do("HSET", "MP_"+multiUploadInfo.UploadId, "filesize", multiUploadInfo.FileSize)

	log.Println("----------------------------初始化信息写入到redis缓存中成功----------------------------")
	return true
}

/*
初始化分块信息Handler
*/
func InitiateMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		log.Println("----------------------------初始化分块上传时文件大小转换为整数失败请稍后再试----------------------------")
		w.Write(util.NewRespMsg(-1, "Param Invalid", nil).JSONBytes())
		return
	}
	multiUploadInfo := MultipartUploadInfo{}
	iniRet := InitializeMulpartInfo(filehash, filesize, username, &multiUploadInfo)
	if !iniRet {
		log.Println("----------------------------初始化分块信息失败，请稍后再试----------------------------")
		w.Write(util.NewRespMsg(-1, "初始化分块信息失败，请稍后再试", nil).JSONBytes())
		return
	}
	//5. 将初始化信息返回给客户端
	w.Write(util.NewRespMsg(0, "OK", multiUploadInfo).JSONBytes())

}

func UploadPart(uploadID string, chunkIndex string, b io.ReadCloser) bool {
	//2. 获得redis连接池的连接
	rConn := rPool.GetRedisConn().Get()
	//不要忘记在退出的时候关闭redis的连接
	defer rConn.Close()

	//3. 根据当前用户以及上传文件块的内容获得文件句柄
	//如果通过create创建文件的时候，目录不存在直接会报错
	//解决方案首先创建目录再创建文件
	fpath := "./data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744) //当前用户拥有7的权限，其他只有4的权限
	fd, err := os.Create(fpath)
	if err != nil {
		log.Println("----------------------------创建存储文件分块内容的文件失败----------------------------")
		return false
	}
	defer fd.Close()

	//将得到文件的分块内容存储起来
	//通过一个for循环每次只读1M或者10M
	buf := make([]byte, 16*1024*1024)
	for {
		n, err := b.Read(buf)
		fd.Write(buf[:n])
		if err != nil { //如果读到文件最后就要退出当前循环，此时err不为nil
			break
		}
	}

	//TODO：(自己看一下加在哪里)还可以完善一下，加一个分块哈希的校验，每次都需要客户端上传本地计算好的哈希值，服务端接收到之后就比较，如果一致说明没有被串改

	//4. 上传文件分块之后更新redis中的缓存数据，表明当前块已经上传完成
	//我们每次完成一个上传分块，就往当前key为Uploadid里面的hashset里面添加一条记录
	//之后便于我们查询上传分块的进度，一旦发现所有分块都上传上去之后就可以合并了
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	log.Println("----------------------------分块上传成功----------------------------")
	return true
}

/*
上传文件分块
*/
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析用户请求参数
	r.ParseForm()
	_ = r.Form.Get("username")
	uploadID := r.Form.Get("uploadid") //说明当前分块是属于哪一个uploadID的请求
	chunkIndex := r.Form.Get("index")  //说明是文件内容的第几块
	upRet := UploadPart(uploadID, chunkIndex, r.Body)

	if !upRet {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}

	//5. 返回处理结果给客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

func CompleteUploadPart(uploadId string, username string, filehash string, filesize string, filename string) bool {
	//2. 获得redis连接池里面的链接
	rConn := rPool.GetRedisConn().Get()
	//不要忘记在退出的时候关闭redis的连接
	defer rConn.Close()

	//3. 通过uploadId查询redis里面的数据，判断是否所有分块都上传完成
	//我们通过hgetall取出当前uploadid里面在redis缓存中的所有数据，如果条数与我们的分块数目一致则说明上传完成
	//需要将查询出的结果转换为interface的array
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadId))
	//查询失败，返回错误信息给客户端
	if err != nil {
		log.Println("----------------------------查询uploadid信息失败----------------------------")
		return false
	}

	totalCount := 0
	chunkCount := 0

	//跳转为2，因为通过hgetall查出来的时候，key和value都在里面
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))

		//获取到了我们初始化写入的chunkcount也就是一共要分成多少块
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" { //如果以chkidx_开头并且对应的值为1说明该分块上传完成
			chunkCount++
		}
	}

	//说明没有上传完，直接返回一个错误信息
	if totalCount != chunkCount {
		log.Println("----------------------------分块上传不完整----------------------------")
		return false
	}

	//4. TTODO：在所有的分块上传完成之后，做一个分块合并的操作（老师暂时先跳过这一步），
	mergeRet := MergeMultiPart(uploadId, totalCount, filename)
	if !mergeRet {
		log.Println("----------------------------分块合并失败----------------------------")
		return false
	}

	//文件分块合并之后，我们需要删除对应的文件夹以及redis中的缓存记录
	//删除对应的临时文件夹
	err = os.RemoveAll("./data/" + uploadId + "/")
	if err != nil {
		log.Println("----------------------------删除uploadid对应的临时文件夹失败，请稍后再试----------------------------")
		return false
	}
	//删除redis中的记录
	_, err = rConn.Do("DEL", "MP_"+uploadId)
	if err != nil {
		log.Println("----------------------------Redis中删除键为uploadid的记录失败，请稍后再试----------------------------")
		return false
	}

	//5. 文件合并完成之后需要更新两个表：文件表，用户文件表
	fsize, err := strconv.Atoi(filesize)
	if err != nil {
		log.Println("----------------------------文件大小转整数失败----------------------------")
		return false
	}

	//自己做完了：TTODO:由于前面的合并我们还没有做，所以这的filelocation我们后面补上
	fMetaData := model.FileMetaData{FileHash: filehash, FileSize: int64(fsize), FileName: filename, FileLocation: "./tmp/" + filename}
	addRet := dblayer.AddFileMetaData(fMetaData)
	updateRet := dblayer.UpdateUserFile(username, fMetaData)
	if !addRet || !updateRet {
		log.Println("----------------------------文件表或用户文件表更新失败----------------------------")
		return false
	}

	return true
}

/*
通知上传合并接口
*/
func CompleteUploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析请求参数
	r.ParseForm()
	uploadId := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	cRet := CompleteUploadPart(uploadId, username, filehash, filesize, filename)
	if !cRet {
		log.Println("----------------------------通知上传合并失败----------------------------")
		w.Write(util.NewRespMsg(-2, "通知上传合并失败", nil).JSONBytes())
		return
	}

	//6. 向客户端响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())

}

/*
取消上传分块
实现思路：
1. 云端接收到取消的请求之后，首先检查UploadId是否有效，如果无效直接返回，如果有效进行如下操作
2. 删除已经存在的分块文件：
3. 删除redis中的缓存状态：根据用户名以及uploadid查到对应的记录，之后删除掉
4. 更新mysql文件的status：这一步不是必须的
*/
func CancelUploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//TODO：老师不打算继续讲解，自己参考老师思路实现
	//1. 解析用户请求参数
	r.ParseForm()
	uploadId := r.Form.Get("uploadid")

	//2. 判断uploadid是否有效
	//获得redis连接池里面的链接
	rConn := rPool.GetRedisConn().Get()
	//不要忘记在退出的时候关闭redis的连接
	defer rConn.Close()

	ret, err := redis.Values(rConn.Do("HLEN", uploadId))
	if err != nil {
		log.Println("----------------------------查询redis中是否存在这个key出错----------------------------")
		w.Write(util.NewRespMsg(-1, "查询redis中是否存在这个key出错", nil).JSONBytes())
		return
	}
	if ret[0].(int) == 0 {
		log.Println("----------------------------redis不存在这个uploadid----------------------------")
		w.Write(util.NewRespMsg(0, "redis不存在这个uploadid", nil).JSONBytes())
		return
	}

	//如果redis中存在这个uploadid
	uploadStatus, err := MultipartUploadStatus(uploadId)
	if err != nil {
		log.Println("----------------------------查看分块上传的整体状态失败----------------------------")
		w.Write(util.NewRespMsg(-1, "查看分块上传的整体状态失败，请稍后再试", nil).JSONBytes())
	}

	//TODO：自己设置一个事件自动删除用户缓存的过期文件块，也就是在有效时间范围内没有将文件全部上传完成

	//3. 如果有效判断uploadid是否完成，
	//如果完成需要删除数据库表中的记录以及删除已经存在的分块文件
	partFilePath := "./data/" + uploadId + "/"

	//TODO:老师课上讲过，成功上传之后无法取消上传
	//因此自己这里实现的是未成功上传之前的删除，也就是只删除redis中的缓存记录与我们保存到data文件夹下upload对应文件夹中的内容
	//if uploadStatus.TotalCount == uploadStatus.CurCount { //删除数据库表中的记录同时删除已经合并后的文件
	//
	//
	//} else { //如果没有完成需要中断
	//
	//}

	//删除已经存在的分块文件
	for i := 1; i <= uploadStatus.TotalCount; i++ {
		//如果在我们已经完成的分块中就删除
		if _, ok := uploadStatus.CompletedChunkIdx[i]; ok {
			os.Remove(partFilePath + strconv.Itoa(i))
		}
	}

	//删除redis中的缓存状态
	_, err = rConn.Do("DEL", "MP_"+uploadId)
	if err != nil {
		log.Println("----------------------------redis中删除对应的uploadID失败----------------------------")
		w.Write(util.NewRespMsg(-2, "redis中删除对应的uploadID失败，请稍后再试", nil).JSONBytes())
	}
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

type UploadStatus struct {
	TotalCount        int
	CurCount          int
	CompletedChunkIdx map[int]bool
	ProgressPercent   string
}

//查看分块上传的整体状态
func MultipartUploadStatus(uploadId string) (UploadStatus, error) {
	//获得redis连接池里面的链接
	rConn := rPool.GetRedisConn().Get()
	//不要忘记在退出的时候关闭redis的连接
	defer rConn.Close()

	//3. 通过uploadId查询redis里面的数据，判断是否所有分块都上传完成
	//我们通过hgetall取出当前uploadid里面在redis缓存中的所有数据，如果条数与我们的分块数目一致则说明上传完成
	//需要将查询出的结果转换为interface的array
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadId))
	//查询失败，返回错误信息给客户端
	if err != nil {
		log.Println("----------------------------查询uploadid信息失败----------------------------")
		return UploadStatus{}, err
	}

	//存储已经完成的文件块的序号
	completedChunkIdx := make(map[int]bool)

	//存储一下文件一共需要多少块
	totalCount := 0

	//跳转为2，因为通过hgetall查出来的时候，key和value都在里面
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))

		//获取到了我们初始化写入的chunkcount也就是一共要分成多少块
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" { //如果以chkidx_开头并且对应的值为1说明该分块上传完成
			idx, _ := strconv.Atoi(k[7:])
			completedChunkIdx[idx] = true
		}
	}
	notCompletedIdx := []int{}
	//1.2 查询之后统计出还没有上传的所有文件块的序号，同时统计进度并进行返回
	for i := 1; i <= totalCount; i++ {
		if _, ok := completedChunkIdx[i]; !ok {
			notCompletedIdx = append(notCompletedIdx, i)
		}
	}

	//返回还没有上传的文件的序号以及一共需要上传的文件块的数目以及目前上传的进度
	return UploadStatus{
		CurCount:          len(completedChunkIdx),
		TotalCount:        totalCount,
		ProgressPercent:   fmt.Sprintf("%.2f", float64(len(completedChunkIdx))/float64(totalCount)),
		CompletedChunkIdx: completedChunkIdx,
	}, nil
}

/*
查看分块上传的整体状态

最核心的：
1. 通过用户名以及uploadid从redis中查到所有还没有上传的文件块状态进行统计封装之后封装返回给用户

假设某一个文件有100个块，但是在redis中只查到50个块，那么客户端看到的进度条应该是50%，然后把还没有上传的序号返回给客户端，之后客户端可以根据还未上传的序号进行上传
*/
func MultipartUploadStatusHandler(w http.ResponseWriter, r *http.Request) {
	//TODO：老师不打算继续讲解，自己参考老师思路实现吧
	//1. 解析用户请求中的表单
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadId := r.Form.Get("uploadid")
	//1.1 此时从redis中查询上传的所有文件块
	uploadStatus, err := MultipartUploadStatus(uploadId)
	if err != nil {
		log.Println("----------------------------查看分块上传的整体状态失败----------------------------")
		w.Write(util.NewRespMsg(-1, "查看分块上传的整体状态失败，请稍后再试", nil).JSONBytes())
		return
	}
	w.Write(util.NewRespMsg(0, "OK", uploadStatus).JSONBytes())
	return
}

//TODO：断点续传，只要将分块上传中断一下，客户端再获取上传进度，之后知道自己还要上传哪些分块，之后继续上传这些分块就可以了，

/*
合并多个文件
第1个参数为要合并文件存放的uploadId，第2个为应该要合并多少个文件，第3个参数表示合并之后的文件名
*/
func MergeMultiPart(uploadId string, totalCount int, filename string) bool {

	//创建1M的缓冲区
	buff := make([]byte, 1024*1024)
	//1. 将多个分块合并："./data/" + uploadID + "/" + chunkIndex
	//分块文件路径
	partFile := "./data/" + uploadId + "/"
	//合并后的文件存放的路径
	dstFile, err := os.Create("./tmp/" + filename)
	if err != nil {
		log.Println("----------------------------创建合并后的文件失败，请稍后再试----------------------------")
		return false
	}
	for i := 1; i <= totalCount; i++ {
		f, err := os.Open(partFile + strconv.Itoa(i))
		if err != nil {
			log.Println("----------------------------打开文件（\"./data/\" + uploadId + \"/\" + strconv.Itoa(i)）失败----------------------------")
			return false
		}
		for {
			writeSize, err := f.Read(buff)
			if err != nil {
				break
			}
			dstFile.Write(buff[:writeSize])
		}
	}

	return true
}
