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
> 使用 readinessProbe探针 检查pod是否就绪， 只有就绪的情况下才接收请求。

- 优雅终止。
> 使用配置 terminationGracePeriodSeconds: 60， 在pod发出关闭指令时，k8s将给应用发送SIGTERM信号，k8s会等待60秒后关闭。
> 
> httpserver源码中增加对SIGTERM信号的检测处理。

- 资源需求和QoS保证。
> 设置deployment的resources请求
> 
> Qos有三种服务质量等级，分别是：
> 
> Guaranteed：Pod 里的每个容器都必须有内存/CPU 限制和请求，而且值必须相等。如果一个容器只指明limit而未设定request，则request的值等于limit值。
> 
> Burstable：Pod 里至少有一个容器有内存或者 CPU 请求且不满足 Guarantee 等级的要求，即内存/CPU 的值设置的不同。
> 
> BestEffort：容器必须没有任何内存或者 CPU 的限制或请求。
> 
> 本次实验由于资源充分，采用 Guaranteed 方式

- 探活。
> 使用 livenessProbe 存活探针进行探测，如果检测到应用没有存活就杀掉当前pod并重启。

- 日常运维需求，日志等级。
> httpserver应用程序使用glog的日志级别，替代golang原生的log包

- 配置和代码分离。
> 使用configMap将常用配置注入到pod中

- 如何确保整个应用的高可用
> 部署 deployment时 增加多个副本 -- replicas: 2

- 如何通过证书保证httpServer的通讯安全
> httpserver应用程序本身并没有增加https证书， 可以在ingress侧增加证书验证，实现认证与代码分离
```yaml
  tls:
    - hosts:
        - dhtobb.com
      secretName: tls-secret #使用tls-secret作为证书
```

##实验环境
```shell
#3台虚拟机， 一台master节点， 两台node节点
root@master:~/hs/specs# kubectl get node -owide
NAME     STATUS   ROLES                  AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION     CONTAINER-RUNTIME
master   Ready    control-plane,master   36d   v1.22.2   192.168.34.2   <none>        Ubuntu 20.04.3 LTS   5.4.0-89-generic   docker://20.10.8
node2    Ready    <none>                 36d   v1.22.2   192.168.34.3   <none>        Ubuntu 20.04.3 LTS   5.4.0-89-generic   docker://20.10.8
node3    Ready    <none>                 28m   v1.22.2   10.252.3.72    <none>        Ubuntu 20.04.3 LTS   5.4.0-89-generic   docker://20.10.8

root@master:~/hs/specs# kubectl get node node2 -oyaml
apiVersion: v1
kind: Node
metadata:
...
spec:
...
status:
...
  allocatable:
    cpu: "2"
    ephemeral-storage: "18903225108"
    hugepages-2Mi: "0"
    memory: 1932828Ki
    pods: "110"
  capacity:
    cpu: "2"
    ephemeral-storage: 20511312Ki
    hugepages-2Mi: "0"
    memory: 2035228Ki
    pods: "110"
...
```

##操作
###1. 创建configmap、pvc以及部署应用
```shell
root@master:~/hs/specs# kubectl apply -f configmap.yaml

root@master:~/hs/specs# kubectl apply -f pv.yaml
root@master:~/hs/specs# kubectl apply -f pvc.yaml

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

#查看configMap 的值是否映射到POD中
root@master:~/hs/specs# kubectl exec -it dhtobb-httpserver-9fb59ccf4-6rxks -- env
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=dhtobb-httpserver-777b486958-hkx5z
LogDir=/hs-log #--- 已经生效
```

###2. 创建service
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

###3. 创建ingress
#### install ingress controller
```
kubectl create -f nginx-ingress-deployment.yaml
```

问题：nginx-ingress-controller启动失败
```
#解决过程：
#查看nginx-ingress-deployment
root@master:~/hs/specs# kubectl get pod --namespace ingress-nginx
NAME                                       READY   STATUS              RESTARTS   AGE
ingress-nginx-admission-create--1-wlq88    0/1     ImagePullBackOff    0          15m
ingress-nginx-admission-patch--1-hv5dx     0/1     ImagePullBackOff    0          15m
ingress-nginx-controller-8cf5559f8-stsh8   0/1     ContainerCreating   0          15m

root@master:~/hs/specs# kubectl describe pod ingress-nginx-admission-create--1-wlq88
#问题原因是镜像拉取失败
Error response from daemon: Get "https://k8s.gcr.io/v2/":

#无法从k8s.gcr.io/v2/上拉取镜像，改为从aliyuncs上拉取镜像，拉取之后直接增加tag， 也可以直接修改yaml文件改为能够访问的镜像
# docker pull registry.aliyuncs.com/google_containers/kube-webhook-certgen:v1.0
# docker tag registry.aliyuncs.com/google_containers/kube-webhook-certgen:v1.0 k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.0
# docker rmi registry.aliyuncs.com/google_containers/kube-webhook-certgen:v1.0
```

#### generated key-cert
```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=dhtobb.com/O=dhtobb"
```
#### 生成secret yaml配置文件，
```
#1、通过openssl生成的key-crt文件生成关于secrets的yml文件
eg: kubectl create secret tls NAME --cert=path/to/cert/file --key=path/to/key/file [--dry-run=server|client|none]

kubectl create secret tls tls-secret --cert=tls.crt --key=tls.key --dry-run=client -o yaml


2、从已经生成的secret导出为yaml文件
kubectl get secret tls-secret -n default -o yaml > secret.yaml

```
#### 创建secret
```shell
#1、直接通过key-crt文件创建
kubectl create secret tls tls-secret --cert=tls.crt --key=tls.key

#2、通过yaml文件创建
kubect create -f secret.yaml
```

#### create a ingress
```
kubectl create -f ingress.yaml
```

问题1: Error from server (InternalError): error when creating "ingress.yaml": Internal error occurred: failed calling webhook "validate.nginx.ingress.kubernetes.io": Post "https://ingress-nginx-controller-admission.ingress-nginx.svc:443/networking/v1/ingresses?timeout=10s": Service Unavailable
root@master:~/hs/specs# kubectl get validatingwebhookconfigurations
NAME                      WEBHOOKS   AGE
ingress-nginx-admission   1          4m16s

解决方案：删除ingress-nginx-adminssion
root@master:~/hs/specs# kubectl delete -A validatingwebhookconfigurations ingress-nginx-admission
validatingwebhookconfiguration.admissionregistration.k8s.io "ingress-nginx-admission" deleted


#### 查看ingress启动情况
```shell
root@master:~/hs/specs# kubectl get ingress
NAME             CLASS    HOSTS        ADDRESS        PORTS     AGE
dhtobb-gateway   <none>   dhtobb.com   192.168.34.3   80, 443   6h58m

root@master:~/hs/specs# kubectl get svc -n ingress-nginx
NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
ingress-nginx-controller             NodePort    10.100.242.30   <none>        80:30566/TCP,443:30531/TCP   55m
ingress-nginx-controller-admission   ClusterIP   10.109.90.140   <none>        443/TCP                      55m
```
#### test the result
```shell
root@master:~/hs/specs# curl --noproxy "*" -H "Host: dhtobb.com" https://10.100.242.30 -v -k

#可通过物理机的IP与端口对集群进行访问
root@master:~/hs/specs# curl --noproxy "*" -H "Host: dhtobb.com" https://10.252.3.70:30531 -k
It works!

Service IP is: 192.168.104.21

#查看日志存储路径是否在configmap中设置的环境变量 /hs-log 目录下
root@master:~/hs/specs# kubectl exec -it dhtobb-httpserver-5bc8ddd969-fcbkp -- ls /hs-log/
httpserver.INFO
httpserver.dhtobb-httpserver-5bc8ddd969-fcbkp.root.log.INFO.20211128-040721.1

#查看日志内容 -- 可以看到请求被记录到了日志中， /healthz 健康探针被kubelet调用
root@master:~/hs/specs# kubectl exec -it dhtobb-httpserver-5bc8ddd969-fcbkp -- cat /hs-log/httpserver.dhtobb-httpserver-5bc8ddd969-fcbkp.root.log.INFO.20211128-040721.1
Log file created at: 2021/11/28 04:07:21
Running on machine: dhtobb-httpserver-5bc8ddd969-fcbkp
Binary: Built with gc go1.17.2 for linux/amd64
Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg
I1128 04:07:21.445540       1 main.go:132] startup server and listen on port... 80
I1128 04:07:34.368148       1 main.go:93] path:  /healthz Client IP:  10.252.3.72:35150 , HTTP Code:  200
I1128 04:08:04.367114       1 main.go:93] path:  /healthz Client IP:  10.252.3.72:35182 , HTTP Code:  200
I1128 04:08:04.367932       1 main.go:93] path:  /healthz Client IP:  10.252.3.72:35180 , HTTP Code:  200
I1128 04:08:34.366938       1 main.go:93] path:  /healthz Client IP:  10.252.3.72:35218 , HTTP Code:  200
I1128 04:08:34.366948       1 main.go:93] path:  /healthz Client IP:  10.252.3.72:35220 , HTTP Code:  200
I1128 04:08:46.672844       1 main.go:93] path:  / Client IP:  192.168.104.23:53060 , HTTP Code:  200

```
