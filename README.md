# DistributedFileSystem
自己实践分布式企业存储


## TODO
1. 下载文件：在文件上传保存之后，在下载那一端部署一个反向代理，然后将文件作为一个静态资源来处理，例如nginx，下载的时候后端服务会提供一个接口，用于构造下载文件的url，客户端获取url之后，就去下载，下载的时候会经过nginx，nginx再做一次静态资源访问将文件download下来，一些限流以及权限访问都可以在nginx做，可以减轻golang实现后端的压力。
2. 


## 第2章节

前面的代码进行纠正：

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

实现了文件的基本功能如下：
1. 上传文件：到`http://localhost:8080/file/upload`    进入页面传入要上传的文件
2. 查询上传文件的元信息(get)：`http://localhost:8080/file/meta?filehash=文件的hash值` 文件的sha1值
3. 查询最近上传的几个文件(get)：`http://localhost:8080/file/query/?limit=一个值`
4. 文件修改(post)(只是修改了map中的映射文件名，实际上物理存储的文件名没修改)：`http://localhost:8080/file/update?filehash=文件的sha1值&filename=新文件名(如111.png)`
5. 文件删除(get)：`http://localhost:8080/file/delete?filehash=文件sha1值`
6. 文件下载(get)：`http://localhost:8080/file/download?filehash=文件sha1值`


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