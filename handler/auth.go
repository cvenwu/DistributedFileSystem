package handler

import (
	"fmt"
	"log"
	"net/http"
)

/**
 * @Author: yirufeng
 * @Email: yirufeng@foxmail.com
 * @Date: 2020/9/24 11:02 上午
 * @Desc: 使用http拦截器进行权限认证
 */

//拦截器，首先执行这段代码进行权限校验，之后才会分发给具体的业务逻辑函数
func HttpInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		log.Println(username, token)
		if len(username) < 3 || !IsTokenValid(token) {
			fmt.Println(username)
			fmt.Println(token)
			fmt.Println(len(token))
			log.Println("访问已经被拦截----------------------------", r.RequestURI)
			//token校验失败则跳转到登录页面
			http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
			return
		}
		//如果没有拦截就直接下发给具体的业务逻辑处理函数
		h(w, r)
	}
}
