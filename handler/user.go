package handler

import (
	"DFS/db"
	"DFS/util"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/24 9:18 上午
 * @Desc: 处理用户相关的逻辑函数
 */

const pwdSalt = "*#890"

//用户注册处理函数
func UserSignUp(w http.ResponseWriter, r *http.Request) {
	//判断用户请求类型
	if r.Method == http.MethodGet {
		content, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			log.Println("----------------------------注册页面打开失败----------------------------")
			w.Write([]byte("注册页面打开失败"))
		}
		w.Write(content)
	} else if r.Method == http.MethodPost {
		//1. 解析用户请求的参数
		r.ParseForm()
		//2. 获取用户输入的用户名以及密码
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if len(username) < 5 || len(password) < 6 {
			w.Write([]byte("FAILED"))
			return
		}

		//3. 向数据库添加记录

		//因为数据库不是明文存密码的，并且不能直接加密存进去，很容易被破解，
		//这里我们使用用户输入的密码与我们自己的盐值拼接，拼接之后加密传送到数据库中

		opRet := db.UserSignUp(username, util.Sha1([]byte(password+pwdSalt)))
		if !opRet {
			log.Println("----------------------------用户注册失败----------------------------")
			w.Write([]byte("FAILED"))
			return
		}
		//4. 返回响应
		w.Write([]byte("SUCCESS"))
	}

}

//用户登录处理函数
func UserSignIn(w http.ResponseWriter, r *http.Request) {
	//如果是get请求，直接传登录页面过去
	if r.Method == http.MethodGet {
		content, err := ioutil.ReadFile("./static/view/signin.html")
		if err != nil {
			log.Println("----------------------------登录页面打开失败----------------------------")
			w.Write([]byte("登录页面打开失败"))
			return
		}
		w.Write(content)
		return

		////http://localhost:8080/user/signin
		//	http.Redirect(w, r, "./static/view/signin.html", http.StatusSeeOther)
		//	return
	}
	//如果是Post请求，说明用户要进行登录
	//1. 解析用户带的用户名以及密码
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	//2. 验证用户的用户名以及密码是否正确
	opRet := db.UserSignIn(username, util.Sha1([]byte(password+pwdSalt)))
	if !opRet {
		log.Println("----------------------------用户名或密码不正确----------------------------")
		w.Write([]byte("FAILED"))
		return
	}

	//3. 如果正确就更新生成的token给用户，之后用户靠这个token访问非注册登录的其他功能
	token := GenToken(username)
	//向用户token更新对应用户之前的记录
	upRet := db.UpdateUserToken(username, token)
	if !upRet {
		log.Println("----------------------------更新用户token失败----------------------------")
		w.Write([]byte("FAILED"))
		return
	}
	log.Println("登录校验认证通过=====================================")
	//4. 之后将我们新生成的token返回给用户,同时重定向的url我们也交给浏览器自己去做
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html", //要跳转到页面的地址
			Username: username,
			Token:    token,
		},
	}
	//返回序列化后的内容
	w.Write(resp.JSONBytes())
}

//用户信息查询接口
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("userinfo接收到一个用户的请求")
	if r.Method == http.MethodPost {

		//1. 解析请求参数
		r.ParseForm()
		username := r.Form.Get("username")
		//token := r.Form.Get("token")
		////2. 验证用户携带的token是否有效:这里的验证我们后面改用了拦截器去做
		//if !IsTokenValid(token) {
		//	log.Println("----------------------------token无效，请稍后再试----------------------------")
		//	w.Write([]byte("token无效，请稍后再试"))
		//	return
		//}

		//3. 如果有效我们直接查询用户信息
		userRet, err := db.GetUserInfo(username)
		if err != nil {
			log.Println("----------------------------获取用户信息失败----------------------------")
			w.Write([]byte("获取用户信息失败，请稍后再试"))
			//w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println(userRet.SignupAt)
		//4. 将查询到的用户信息封装返回给客户
		resp := util.RespMsg{
			Code: 0,
			Msg:  "OK",
			Data: userRet,
		}
		w.Write(resp.JSONBytes())
	}
}

//生成token
//token生成格式：md5 对字符串(username + 时间戳 + _tokensalt)加密
//token我们一共要拼成40位：md5生成的是32位加密后的内容，之后8位我们采用时间戳的前8位
func GenToken(username string) string {
	//%X	表示为十六进制，使用A-F
	timeStamp := fmt.Sprintf("%X", time.Now().Unix())
	token := util.MD5([]byte(username+timeStamp)) + timeStamp[:8]
	return token
}

//这里我们只是简单的验证一下token是否为我们自己设置的md5的长度40位
func IsTokenValid(token string) bool {
	//TODO:首先验证是否符合我们的语法规范，然后验证是否在有效时间内，然后参考老师课上讲解的验证token的有效思路来进行验证
	if len(token) != 40 {
		return false
	}
	return true
}
