package redis

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/26 9:43 上午
 * @Desc:
 */

import (
	"DFS/config"
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
)

var (
	//连接池对象
	pool *redis.Pool
)

//创建Redis的连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   50,
		MaxActive: 30,
		//5分钟
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			//1. 打开连接
			conn, err := redis.Dial("tcp", config.RedisHost)
			if err != nil {
				log.Println("----------------------------redis连接初始化失败----------------------------")
				return nil, err
			}
			//2. 访问认证：如果有密码要输入密码
			//conn.Do("AUTH", config.RedisPass)
			log.Println("----------------------------redis连接初始化成功----------------------------")

			return conn, nil
		},
		//用于定时检查连接是否可用，如果出状况直接在客户端关闭redis连接
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			//每分钟检测一次redis-server的可用性
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

//初始化的时候就给他赋值
func init() {
	pool = newRedisPool()
}

//对外暴露一个接口来获取redis的连接
func GetRedisConn() *redis.Pool {
	return pool
}
