##课后练习4.2：

---

- 启动一个 Envoy Deployment。 
- 要求 Envoy 的启动配置从外部的配置文件 Mount 进 Pod。 
- 进入 Pod 查看 Envoy 进程和配置。 
- 更改配置的监听端口并测试访问入口的变化。 
- 通过非级联删除的方法逐个删除对象。

##操作步骤

---
0-0. 准备envoy的部署配置文件envoy-deploy.yaml

0-1. 准备envoy的配置文件envoy.yaml

1. 创建configmap
```shell
root@master:~/cn# kubectl create configmap envoy-config --from-file=envoy.yaml
configmap/envoy-config created
```

2. 启动Envoy Deployment
```shell
root@master:~/cn# kubectl create -f envoy-deploy.yaml
deployment.apps/envoy created
``` 

2. 创建envoy的service，暴露端口
```shell
 root@master:~/cn# kubectl expose deploy envoy --selector run=envoy --port=10000 --target-port=10000 --type=NodePort
service/envoy exposed


root@master:~/cn# kubectl get svc -owide
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)           AGE    SELECTOR
envoy        NodePort    10.98.160.227   <none>        10000:32144/TCP   45m    run=envoy

```

3. 查看pod部署的节点位置
```shell
root@master:~/cn# kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE   IP              NODE    NOMINATED NODE   READINESS GATES
envoy-fb5d77cc9-p5f8g   1/1     Running   0          27m   192.168.104.4   node2   <none>           <none>
   
#目前pod状态为ContainerCreating,查看pod的详细描述
root@master:~/cn# kubectl describe pod envoy-fb5d77cc9-p5f8g
Name:         envoy-fb5d77cc9-p5f8g
Namespace:    default
Priority:     0
Node:         node2/192.168.34.3
Start Time:   Tue, 26 Oct 2021 08:31:33 +0000
Labels:       pod-template-hash=fb5d77cc9
              run=envoy
Annotations:  cni.projectcalico.org/containerID: 262488a5dd8437e08d61acd2dd239251a9a8bdbe7567578901fec9e7bdbffbe7
              cni.projectcalico.org/podIP: 192.168.104.4/32
              cni.projectcalico.org/podIPs: 192.168.104.4/32
Status:       Running
IP:           192.168.104.4
IPs:
  IP:           192.168.104.4
Controlled By:  ReplicaSet/envoy-fb5d77cc9


```


3.访问服务
```shell
root@master:~/cn# curl --noproxy "*" 192.168.104.4:10000
no healthy upstreamr
``` 

4.进入 Pod 查看 Envoy 进程和配置。
```shell
root@master:~/cn# kubectl exec -it envoy-fb5d77cc9-p5f8g -- bash

root@envoy-fb5d77cc9-p5f8g:/# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
envoy          1       0  0 08:43 ?        00:00:03 envoy -c /etc/envoy/envoy.yaml
root          19       0  0 09:05 pts/0    00:00:00 bash
root          30      19  0 09:05 pts/0    00:00:00 ps -ef

root@envoy-fb5d77cc9-p5f8g:/# cat /etc/envoy/envoy.yaml

```
5.更改配置的监听端口并测试访问入口的变化。 
```shell
root@master:~/cn# kubectl edit svc/envoy
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
#
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2021-10-26T08:32:13Z"
  labels:
    run: envoy
  name: envoy
  namespace: default
  resourceVersion: "105839"
  uid: a7669ffb-1fc6-40dd-bfc9-068cdd49c619
spec:
  clusterIP: 10.98.160.227
  clusterIPs:
  - 10.98.160.227
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - nodePort: 32144
    port: 20000 #change 10000 to 20000
    protocol: TCP
    targetPort: 10000
  selector:
    run: envoy
  sessionAffinity: None
  type: NodePort
status:
  loadBalancer: {}


root@master:~/cn# kubectl get svc
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)           AGE
envoy        NodePort    10.98.160.227   <none>        20000:32144/TCP   57m


```

6.通过非级联删除的方法逐个删除对象。
```shell
#查看deployment以及po的情况
root@master:~/cn# kubectl get deploy -owide
NAME    READY   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS   IMAGES                 SELECTOR
envoy   1/1     1            1           86m   envoy        envoyproxy/envoy-dev   run=envoy
root@master:~/cn# kubectl get po -owide
NAME                    READY   STATUS    RESTARTS   AGE    IP              NODE    NOMINATED NODE   READINESS GATES
envoy-fb5d77cc9-p5f8g   1/1     Running   0          87m    192.168.104.4   node2   <none>           <none>

#删除deployment
root@master:~/cn# kubectl delete deploy  envoy --cascade=orphan
deployment.apps "envoy" deleted

#此时pod/service/configMap/replicasets依然存在
root@master:~/cn# kubectl get po
NAME                    READY   STATUS    RESTARTS   AGE
envoy-fb5d77cc9-p5f8g   1/1     Running   0          91m

root@master:~/cn# kubectl get cm
NAME               DATA   AGE
envoy-config       1      94m

root@master:~/cn# kubectl get svc
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)           AGE
envoy        NodePort    10.98.160.227   <none>        20000:32144/TCP   91m

root@master:~/cn# kubectl get rs
NAME              DESIRED   CURRENT   READY   AGE
envoy-fb5d77cc9   1         1         1       94m

#删除replicasets
root@master:~/cn# kubectl delete rs envoy-fb5d77cc9 --cascade=orphan
replicaset.apps "envoy-fb5d77cc9" deleted

#删除pod
root@master:~/cn# kubectl delete po envoy-fb5d77cc9-p5f8g
pod "envoy-fb5d77cc9-p5f8g" deleted

#删除service
root@master:~/cn# kubectl delete svc envoy
service "envoy" deleted

#删除configMap
root@master:~/cn# kubectl delete cm envoy-config
configmap "envoy-config" deleted

```

7.级联删除
```shell
root@master:~/cn# kubectl delete deploy envoy
deployment.apps "envoy" deleted

#不指定 --cascade=orphan, 删除deploy的同时会删除replicasets(rs)以及相关pod
#不会删除service以及configMap

root@master:~/cn# kubectl delete cm,svc --selector run=envoy

#删除所有pod
root@master:~/cn# kubectl delete pods --all
```

8. scale up/down/failover
```
# kubectl scale deploy <deployment-name> --replicas=<n>

kubectl scale deploy envoy --replicas=2
```