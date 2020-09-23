package util

import (
	"DFS/model"
	"time"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/23 9:24 下午
 * @Desc: 将文件按照上传时间排序
 */

const baseFormat = "2006-01-02 15:04:05"

type UploadTimeAt []model.FileMetaData

func (a UploadTimeAt) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

//按照时间最近的排在前面
func (a UploadTimeAt) Less(i, j int) bool {
	time1, _ := time.Parse(baseFormat, a[i].UploadTimeAt)
	time2, _ := time.Parse(baseFormat, a[j].UploadTimeAt)
	return time1.UnixNano() > time2.UnixNano()
}

func (a UploadTimeAt) Len() int {
	return len(a)
}
