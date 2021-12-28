package main

import (
	"bytes"
	"cncamp/dhtobb/v1/metrics"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	BindPort            = 80
	GlogAlsologtostderr = false
	GlogLogdir          = ""
)

//go:embed favicon.ico
var favicon []byte

// 全局路由
var routings = map[string]func(w http.ResponseWriter, r *http.Request) ([]byte, int){
	"/":            rootHandler,
	"/favicon.ico": faviconHandler,
	"/healthz":     healthzHandler,
	"/delay":       delayHandler,
}

func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

//根路由
func rootHandler(w http.ResponseWriter, r *http.Request) ([]byte, int) {
	return []byte("It works!\r\n\r\nService IP is: " + GetLocalIp() + "\r\n"), http.StatusOK
}

//当访问 localhost/healthz 时，应返回200
func healthzHandler(w http.ResponseWriter, r *http.Request) ([]byte, int) {
	return []byte("200"), http.StatusOK
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

//随机延时0-2秒
func delayHandler(w http.ResponseWriter, r *http.Request) ([]byte, int) {
	timer := metrics.NewTimer()
	defer timer.ObserveTotal()

	rand.Seed(time.Now().UnixNano())
	delay := randInt(0, 2000)

	time.Sleep(time.Millisecond * time.Duration(delay))

	var display bytes.Buffer
	display.Write([]byte("\r\n\r\nService IP is: " + GetLocalIp() + "\r\n"))
	display.WriteString("\r\n")
	display.Write([]byte(fmt.Sprintf("delay %d milliseconds", delay)))
	display.WriteString("\r\n")

	return display.Bytes(), http.StatusOK
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
	glog.Infoln("path: ", r.RequestURI, "Client IP: ", r.RemoteAddr, ", HTTP Code: ", resCode)
}

func GetParamFromENV() {
	if p, b := os.LookupEnv("BindPort"); b {
		BindPort, _ = strconv.Atoi(p)
	}

	if p, b := os.LookupEnv("Alsologtostderr"); b {
		GlogAlsologtostderr, _ = strconv.ParseBool(p)
	}

	if p, b := os.LookupEnv("LogDir"); b {
		GlogLogdir = p
	}
}

func main() {
	GetParamFromENV()

	_ = flag.Set("alsologtostderr", strconv.FormatBool(GlogAlsologtostderr))
	_ = flag.Set("log_dir", GlogLogdir)
	flag.Parse()
	defer func() {
		glog.Flush()
	}()

	//注册prometheus指标
	metrics.Register()

	mux := http.NewServeMux()
	mux.HandleFunc("/", DistributeHandler)
	//prometheus
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", BindPort),
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		glog.Infoln("startup server and listen on port...", BindPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			glog.Fatalln("http server error, ", err.Error())
		}
	}()

	<-quit

	glog.Infoln("shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		glog.Errorln("Server shutdown: ", err)
	} else {
		glog.Infoln("Server has been withdrawn")
	}
}
