# DistributedFileSystem
![](https://img.shields.io/badge/build-passing-brightgreen)

> 一个基于Go的分布式网盘

## 实现的功能

实现了文件的基本功能如下：
1. 上传文件：到`http://localhost:8080/file/upload`    进入页面传入要上传的文件
2. 查询上传文件的元信息(get)：`http://localhost:8080/file/meta?filehash=文件的hash值` 文件的sha1值
3. 查询最近上传的几个文件(get)：`http://localhost:8080/file/query/?limit=一个值`
4. 文件修改(post)(只是修改了map中的映射文件名，实际上物理存储的文件名没修改)：`http://localhost:8080/file/update?filehash=文件的sha1值&filename=新文件名(如111.png)`
5. 文件删除(get)：`http://localhost:8080/file/delete?filehash=文件sha1值`
6. 文件下载(get)：`http://localhost:8080/file/download?filehash=文件sha1值`
7. 用户的注册：`http://localhost:8080/user/signup`
8. 登录：`http://localhost:8080/user/signin`
9. 用户的其他各种操作(需要权限校验)：进行上述与文件的相关操作
10. 文件的分块上传与合并：代码已实现，还没有集成到整个流程中



## TODO
1. 秒传----->使用redis判断是否可以秒传
2. 用户上传之后的文件名以及文件哈希存在redis里面
3. 下载文件：在文件上传保存之后，在下载那一端部署一个反向代理，然后将文件作为一个静态资源来处理，例如nginx，下载的时候后端服务会提供一个接口，用于构造下载文件的url，客户端获取url之后，就去下载，下载的时候会经过nginx，nginx再做一次静态资源访问将文件download下来，一些限流以及权限访问都可以在nginx做，可以减轻golang实现后端的压力。



## 第2章节

老师的代码进行纠正：

之前写的代码：
```go
//理论上工作是已经做完了的，但是为了让浏览器做一个演示，我们需要将一个http的响应头，让浏览器识别出来就可以当成一个文件进行下载
w.Header().Set("Content-Type", "application/octect-stream")
w.Header().Set("Content-Description", "attachment;filename=\""+fm.FileName+"\"")
```


纠正后的代码：
```go
//理论上工作是已经做完了的，但是为了让浏览器做一个演示，我们需要将一个http的响应头，让浏览器识别出来就可以当成一个文件进行下载
w.Header().Set("Content-Type", "application/octect-stream")
w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMetaData.FileName))
log.Println("响应头设置成功....................")
```



## 第4章节
问题描述：
1. 注册登录之后，点击下载，显示一堆乱码，后来看了一下后台返回的是用户的内容，而我们的Js则是读取了后台返回的内容，但是我们的js应该是向我们的下载文件对应的url发起请求触发文件下载
2. 从数据库中查询的结果在网页显示：2020-09-25T18:38:34+08:00 后来发现是数据库存储的问题，解决方法就是在我们在对应逻辑函数中使用sql语句查询的时候就对日期进行格式化 
```sql
select DATE_FORMAT(update_at,'%Y-%m-%d %H:%i:%S') from tbl_file where id = 1
```
业务逻辑处理修改示例：`stmt, err := mysql.GetMysqlConn().Prepare("Select file_sha1, file_name, file_size, file_addr, create_at, DATE_FORMAT(update_at,'%Y-%m-%d %H:%i:%S') from tbl_file where status = 1 order by update_at desc limit ?")`
3. 针对文件大小在网页显示的时候只显示文件以字节为单位的大小，并且只显示没带单位的数字，我们自己做了如下优化
在util下自己百度到了一个文件大小格式化的单位，将以字节为单位的大小的int64输入进去，输出一个带单位的字符串，
我们在model/file.go中为文件专门加了一个文件大小的格式化后的字段，专门用于表示文件大小格式化后的单位，
最后我们在home.html中显示文件大小的时候，改为我们上面刚为文件新加入的格式化文件大小的显示字段

[函数参考](https://blog.csdn.net/gaoluhua/article/details/104591055)
```go
// 字节的单位转换 保留两位小数
func FormatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
```


mysql数据库如果自己插入时间记录还是会早8个小时，因此我们插入的时候自己在sql语句后面使用time.Now()对其进行赋值

## 第5章节：秒传
[知乎大神关于如何实现秒传](https://www.zhihu.com/question/20489900)

老师课上的一个衍生的问题：假如有这么一个应用场景，同时有多个用户同时上传同一个文件，这个时候云端逻辑如何处理呢？
> 参考思路：解决方法不唯一，只有结合实际的应用场景才行。

方法：
1. 允许不同用户同时上传同一个文件
2. 先完成上传的先入库
3. 后上传的只更新用户文件表，并且删除已经上传的文件
     
     

## 第6章节 分块上传与断点续传

自己在之前的代码做了几个改进：
1. 用户上传文件之后，不仅插入数据库表中，还会将文件的hash以及名字缓存到redis中，之后其他用户传文件的时候判断是否可以秒传
2. 用户上传文件的时候，首先判断是否可以秒传(从redis的缓存中查找)，如果可以，就直接秒传
                                                      如果系统重启：就将mysql中文件表的所有的文件信息缓存到redis中
3.          如果不满足秒传，那么此时就会进行真正的上传，如果用户上传的文件没有超过我们配置文件中需要分块的阈值，我们就可以直接上传
                                                如果文件超过了我们自己配置文件中需要分块的阈值，就要进行分块上传
4. 

还可以改进的一点：如果上传成功，我们需要将缓存在redis中的记录删除，同时我们要将保存在data文件夹下的对应uploadid文件夹删除

自己最后可以改进的一点：分块上传可以并发进行，即都可以无序传输，都传上去之后我们就可以对文件进行合并了


自己改进的上传文件，目前还没有完成，会判断是否秒传，是否要分块传输
```go
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
		r.ParseForm()
		username := r.Form.Get("username")


		//1.1 首先判断是否可以秒传

		//1.1.1 读取文件内容计算哈希
		f.Read()
		util.FileSha1()

		//1.1.2 判断是否在redis缓存中
		rConn := rPool.GetRedisConn().Get()
		defer rConn.Close()

		data, err := redis.Values(rConn.Do("HGET tbl_file", "tbl_file:"+filehash))
		if err != nil {
			log.Println("--------------------------从redis中查询文件hash是否在里面失败----------------------------")
			w.Write([]byte("从redis中查询文件hash是否在里面失败"))
			return
		}
		//1.1.3 如果在redis中缓存，则执行秒传
		if len(data[0].(string)) > 0 {

		}




		//1.1.4 如果不在redis缓存中，我们执行下面的完整的上传


		var upRet bool //上传结果
		//如果小于等于我们要分块的阈值，就直接上传
		if fHeader.Size <= config.ChunkThreshold {
			upRet = UploadFile(f, fHeader.Filename, username)
			if !upRet {
				io.WriteString(w, "上传失败，请稍后再试")
				return
			}
		} else {
			//TODO：这里开始整合一下就可以了
			//走到这里说明要分块了
			//1. 初始化分块信息
			InitializeMulpartInfo()
			//2. 上传分块
			UploadPart()
			//3. 合并分块
			MergeMultiPart()

		}

		//重定向：返回一个302响应码，要求用户向响应头部中的url重新发起请求
		http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
	}
}

```



学到的知识：
Golang中删除文件：`os.Remove(文件名)`
Golang中删除空文件夹：`os.Remove(空文件夹名)`
Golang中删除有文件的文件夹：`os.RemoveAll()`


自己测试的文件合并：
```go
package main

import (
	"fmt"
	"log"
	"os"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/26 11:28 上午
 * @Desc:
 */

func main() {


	f, err := os.Open("./蔡徐坤.gif")

	buf := make([]byte, 1024*1024)
	wSize, err := f.Read(buf)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(wSize)

	buf2 := make([]byte, 1024*1024)
	wSize, err = f.Read(buf2)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(wSize)

	buf3 := make([]byte, 1024*1024)
	wSize, err = f.Read(buf3)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(wSize)

	content := []byte{}
	content = append(content, buf...)
	content = append(content, buf2...)
	content = append(content, buf3[:wSize]...)

	newF, err := os.Create("1.gif")
	newF.Write(content)
	newF.Close()
}
```