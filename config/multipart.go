package config

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/26 11:09 上午
 * @Desc:
 */


//分块相关的配置信息
const (
	ChunkSize = 16 * 1024 * 1024  //设置每一个分块大小为16MB
	ChunkThreshold = 16 * 1024 * 1024  //多大的文件就需要分块了
)
