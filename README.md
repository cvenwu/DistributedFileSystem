# DistributedFileSystem
自己实践分布式企业存储


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