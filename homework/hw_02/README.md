##1009模块三作业：

---

- 构建本地镜像。
- 编写 Dockerfile 将练习 2.2 编写的 httpserver 容器化（请思考有哪些最佳实践可以引入到 Dockerfile 中来）。
- 将镜像推送至 Docker 官方镜像仓库。
- 通过 Docker 命令本地启动 httpserver。
- 通过 nsenter 进入容器查看 IP 配置。

##操作步骤

---
0. 编写Dockerfile文件, 采用多段构建
```dockerfile
FROM golang:1.17.2-alpine3.14 AS build

WORKDIR /go/src/httpserver/

COPY httpserver/* /go/src/httpserver/
RUN go env -w GO111MODULE=auto && go build -o /bin/httpserver

FROM alpine:3.13.6
COPY --from=build /bin/httpserver /bin/
EXPOSE 8080
ENTRYPOINT ["/bin/httpserver"]

```
1. 构建本地镜像
> docker build  -t dhtobb/httpserver:v1.0 .

2. 运行镜像
> docker run -it --rm -P --name httpserver  dhtobb/httpserver:v1.0

3.通过 nsenter 进入容器查看 IP 配置
```shell
3.1 查看容器的pid
root@xx:~# docker inspect httpserver | grep -i pid
            "Pid": 27747,
            "PidMode": "",
            "PidsLimit": null,

3.2 查看容器进程的命名空间
root@xx:~# ls -la /proc/27747/ns/
total 0
dr-x--x--x 2 root root 0 Oct 15 09:08 .
dr-xr-xr-x 9 root root 0 Oct 15 09:08 ..
lrwxrwxrwx 1 root root 0 Oct 15 09:08 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 root root 0 Oct 15 09:08 ipc -> 'ipc:[4026532756]'
lrwxrwxrwx 1 root root 0 Oct 15 09:08 mnt -> 'mnt:[4026532754]'
lrwxrwxrwx 1 root root 0 Oct 15 09:08 net -> 'net:[4026532759]'
lrwxrwxrwx 1 root root 0 Oct 15 09:08 pid -> 'pid:[4026532757]'
lrwxrwxrwx 1 root root 0 Oct 15 09:15 pid_for_children -> 'pid:[4026532757]'
lrwxrwxrwx 1 root root 0 Oct 15 09:08 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0 Oct 15 09:08 uts -> 'uts:[4026532755]'
     

3.3 nsenter 进入容器查看 IP 配置
root@kb:~# nsenter -t 27747 -n ip addr
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
105: eth0@if106: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever

```

4. 推送镜像到docker hub官方镜像
```shell
4.1 注册并登录docker hub
docker login

docker push dhtobb/httpserver:v1.0
```