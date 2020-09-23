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