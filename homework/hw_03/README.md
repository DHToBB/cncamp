##1128模块八作业要求：

---
1.编写kubernetes部署脚本将httpserver部署到kubernetes集群。
- 优雅启动。
- 优雅终止。
- 资源需求和QoS保证。
- 探活。
- 日常运维需求，日志等级。
- 配置和代码分离。

2.将服务发布到对内和对外的调用方。
- Service
- Ingress

3.考虑细节：
- 如何确保整个应用的高可用
- 如何通过证书保证httpServer的通讯安全

##分析

---
- 优雅启动。
> 使用 探针 检查pod是否就绪， 只有就绪的情况下才接收请求。

- 优雅终止。
> 使用配置 terminationGracePeriodSeconds: 60， 在pod发出关闭指令时，k8s将给应用发送SIGTERM信号，k8s会等待60秒后关闭。 
> httpserver源码中增加对SIGTERM信号的检测处理。

- 资源需求和QoS保证。

- 探活。

- 日常运维需求，日志等级。
> httpserver应用程序使用glog的日志级别，替代golang原生的log包

- 配置和代码分离。
> 使用configMap将常用配置注入到pod中

##实验环境
```shell
#3台虚拟机， 一台master节点， 两台node节点
root@master:~/hs/specs# kubectl get node -owide
NAME     STATUS   ROLES                  AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION     CONTAINER-RUNTIME
master   Ready    control-plane,master   36d   v1.22.2   192.168.34.2   <none>        Ubuntu 20.04.3 LTS   5.4.0-89-generic   docker://20.10.8
node2    Ready    <none>                 36d   v1.22.2   192.168.34.3   <none>        Ubuntu 20.04.3 LTS   5.4.0-89-generic   docker://20.10.8
node3    Ready    <none>                 28m   v1.22.2   10.252.3.72    <none>        Ubuntu 20.04.3 LTS   5.4.0-89-generic   docker://20.10.8
```

##操作
1. 部署应用
```shell
root@master:~/hs/specs# kubectl apply -f deployment.yaml

#查看pod
root@master:~/hs/specs# kubectl get pod -owide
NAME                                READY   STATUS    RESTARTS   AGE   IP               NODE    NOMINATED NODE   READINESS GATES
dhtobb-httpserver-9fb59ccf4-6rxks   1/1     Running   0          74s   192.168.104.21   node2   <none>           <none>
dhtobb-httpserver-9fb59ccf4-ljkkg   1/1     Running   0          74s   192.168.135.5    node3   <none>           <none>

#通过pod IP访问
root@master:~/hs/specs# curl --noproxy "*" 192.168.135.5
It works!

Service IP is: 192.168.135.5
```

2. 创建service
```shell
root@master:~/hs/specs# kubectl apply -f service.yaml

#查看service
root@master:~/hs/specs# kubectl get svc -owide
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)           AGE   SELECTOR
dhtobb-httpserver-service   ClusterIP   10.99.21.237   <none>        80/TCP            8s    app=httpserver

#访问service
root@master:~/hs/specs# curl --noproxy "*" 10.99.21.237
It works!

Service IP is: 192.168.104.21

```

3. 创建ingress
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