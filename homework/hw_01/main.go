package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed favicon.ico
var favicon []byte

// 全局路由
var routings = map[string]func(w http.ResponseWriter, r *http.Request) ([]byte, int){
	"/":            rootHandler,
	"/favicon.ico": faviconHandler,
	"/healthz":     healthzHandler,
}

//根路由
func rootHandler(w http.ResponseWriter, r *http.Request) ([]byte, int) {
	return []byte("It works!"), http.StatusOK
}

//当访问 localhost/healthz 时，应返回200
func healthzHandler(w http.ResponseWriter, r *http.Request) ([]byte, int) {
	return []byte("200"), http.StatusOK
}

//响应/favicon.ico
func faviconHandler(w http.ResponseWriter, r *http.Request) ([]byte, int) {
	return favicon, http.StatusOK
}

func DistributeHandler(w http.ResponseWriter, r *http.Request) {
	var output []byte
	var resCode int

	callback, ok := routings[r.RequestURI]
	if !ok {
		output = []byte("404")
		resCode = http.StatusNotFound
	} else {
		//1. 接收客户端 request，并将 request 中带的 header 写入 response header
		for k, v := range r.Header {
			w.Header().Add(k, strings.Join(v, ","))
		}

		//2. 读取当前系统的环境变量中的 VERSION 配置，并写入 response header
		w.Header().Add("version", os.Getenv("VERSION"))

		//调用路由单独处理
		output, resCode = callback(w, r)
	}

	//输出响应
	w.WriteHeader(resCode)
	_, _ = w.Write(output)

	//3. Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
	log.Println("Client IP: ", r.RemoteAddr, ", HTTP Code: ", resCode)
}

func main() {
	http.HandleFunc("/", DistributeHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln(err.Error())
	}
}
