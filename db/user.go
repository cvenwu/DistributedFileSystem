package db

import (
	"DFS/db/mysql"
	"DFS/model"
	"log"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/24 9:17 上午
 * @Desc: 用户鉴权系统相关操作
 */

//处理用户注册
func UserSignUp(username string, password string) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("insert into tbl_user(user_name, user_pwd, status) values (?, ?, 1)")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return false
	}

	//将时间转换为字符串
	ret, err := stmt.Exec(username, password)
	if err != nil {
		log.Println("----------------------------用户表插入失败----------------------------", err)
		return false
	}

	if num, err := ret.RowsAffected(); nil == err {
		if num > 0 {
			log.Println("----------------------------注册成功----------------------------")
			return true
		}
		log.Println("----------------------------用户已存在，请重新更换用户名----------------------------")
		return false
	}

	return false
}

//处理用户登录
func UserSignIn(username string, password string) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("select user_name from tbl_user where user_name = ? and user_pwd = ?")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------", err)
		return false
	}

	ret := stmt.QueryRow(username, password)
	if ret != nil {
		log.Println("----------------------------查询到对应用户----------------------------")
		return true
	}

	return false
}

//更新用户表的token
func UpdateUserToken(username string, token string) bool {
	stmt, err := mysql.GetMysqlConn().Prepare("replace into tbl_user_token(user_name, user_token) values (?, ?)")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return false
	}

	ret, err := stmt.Exec(username, token)
	if err != nil {
		log.Println("----------------------------更新token语句执行失败----------------------------")
		return false
	}

	numAffect, err := ret.RowsAffected()

	if nil == err && numAffect > 0 {
		log.Println("----------------------------更新token成功----------------------------")
		return true
	}

	log.Println("----------------------------更新token语句失败----------------------------")
	return false
}

//根据用户名查询用户的信息
func GetUserInfo(username string) (model.User, error) {
	stmt, err := mysql.GetMysqlConn().Prepare("select user_name, DATE_FORMAT(signup_at,'%Y-%m-%d %H:%i:%S') from tbl_user where user_name = ?")
	defer stmt.Close()
	if err != nil {
		log.Println("----------------------------预编译失败----------------------------")
		return model.User{}, err
	}

	//最终返回的查询结果
	userRet := model.User{}

	err = stmt.QueryRow(username).Scan(&userRet.Username, &userRet.SignupAt)
	if err != nil {
		log.Println("----------------------------赋值用户信息结构体失败----------------------------")
		return model.User{}, err
	}

	return userRet, nil
}
