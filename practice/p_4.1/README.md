## 课后练习4.1：

---

- 用 Kubeadm 安装 Kubernetes 集群

## virtual box基本配置

---

1. virtual box虚拟机基本配置要求

> - A compatible Linux host. The Kubernetes project provides generic instructions for Linux distributions based on Debian and Red Hat, and those distributions without a package manager. -- 必须是linux发行版
> - 2 GB or more of RAM per machine (any less will leave little room for your apps). -- 最低2G内存
> - 2 CPUs or more. -- CPU最低2核
> - Full network connectivity between all machines in the cluster (public or private network is fine). -- 所有节点必须网络相同
> - Unique hostname, MAC address, and product_uuid for every node. See [here](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#verify-mac-address) for more details.
> - Certain ports are open on your machines. See [here](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#check-required-ports) for more details.
> - Swap disabled. You **MUST** disable swap in order for the kubelet to work properly.

2. 更改网络配置
```yaml
#vi /etc/netplan/00-installer-config.yaml

network:
    ethernets:
        enp0s3: #根据实际情况设置，用于外部访问虚拟机
          dhcp4: true
        enp0s8: #用于k8s cluster广播地址
          dhcp4: no
          addresses:
            - 192.168.34.2/24
    version: 2
```
```shell
netplan apply  #重启网络
```

3. 关闭swap
```shell
swapoff -a

vi /etc/fstab
#注释掉含有swap关键字的行
```

4. set no password for sudo
```shell
visudo

#添加一行
%sudo ALL=(ALL:ALL) NOPASSWD:ALL
```

## 安装配置容器运行时 docker
1. 安装docker
```shell
apt install docker.io
```

2. 修改配置
```shell
root@node2:~# cat /etc/docker/daemon.json
{
   "registry-mirrors": ["https://xxxx.mirror.aliyuncs.com"], #注册阿里云账号获得加速
   "exec-opts": ["native.cgroupdriver=systemd"] #修改docker cgroup驱动为systemd
}
```

3. 重启docker
```shell
systemctl daemon-reload
systemctl restart docker
```

## install k8s by kubeadm
1. 设置iptables bridged traffic
```shell
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF

# 查看是否生效
root@bb:~# sysctl --system | grep k8s
* Applying /etc/sysctl.d/k8s.conf ...


```

2. 安装kubernets所需的基础工具
```shell
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl
```

3. install kubeadm

```shell
方法1: 在线安装
sudo curl -s https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | sudo apt-key add -

方法2: 离线安装
root@bb:~/work# apt-key add apt-key.gpg
OK
``` 

4. 添加kubernets的阿里镜像源

```shell
sudo tee /etc/apt/sources.list.d/kubernetes.list <<-'EOF'
deb https://mirrors.aliyun.com/kubernetes/apt kubernetes-xenial main
EOF
```

5. 安装kubelet、kubeadm、kubectl

```shell
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl # apt-mark使得软件不会被自动更新
```

6. 初始化集群

```shell
kubeadm init \
 --image-repository registry.aliyuncs.com/google_containers \
 --kubernetes-version v1.22.2 \
 --pod-network-cidr=192.168.0.0/16 \
 --apiserver-advertise-address=192.168.34.2
 
# init操作只需要在master节点上执行即可, 执行完成之后安装提示进行操作即可，重要的是join集群的方法
...
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 192.168.34.2:6443 --token 5aa66s.njfd6gkucihgt8it --discovery-token-ca-cert-hash sha256:34f9a9606733f97e14d7bac859054c81527fff0e3de4b968386afe11a934b75a
``` 

7. 拷贝kubeconfig配置文件。

```shell
  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

```

8. 安装calico 容器网络接口组件
> https://docs.projectcalico.org/getting-started/kubernetes/quickstart
```shell

kubectl create -f https://docs.projectcalico.org/manifests/tigera-operator.yaml
kubectl create -f https://docs.projectcalico.org/manifests/custom-resources.yaml

# 直到所有的pod都处于running状态表示安装完成
watch kubectl get pods -n calico-system

```

9. 去除master节点的污点 untaint master,用以使master节点也可以调度pod部署

```shell
$ kubectl taint nodes --all node-role.kubernetes.io/master-
```

10. 查看node运行情况
```shell
kubectl get nodes -owide
```

11. slave节点加入cluster
> slave节点只要安装到第5步之后，安装calico插件，就可以执行该命令
> 安装calico插件时要用到kubectl，需要使用到master节点的 $HOME/.kube/config 文件，直接拷贝过来即可
> scp <master>:$HOME/.kube/config $HOME/.kube
```shell
kubeadm join 192.168.34.2:6443 --token 5aa66s.njfd6gkucihgt8it --discovery-token-ca-cert-hash sha256:34f9a9606733f97e14d7bac859054c81527fff0e3de4b968386afe11a934b75a

#如果时间相差太久，token过期，则可以重启生成token
kubeadm token list

#生成新的token并打印显示join命令
root@master:~# kubeadm token create --print-join-command
kubeadm join 192.168.34.2:6443 --token dcschc.9qmcqla6edj9uaf6 --discovery-token-ca-cert-hash sha256:34f9a9606733f97e14d7bac859054c81527fff0e3de4b968386afe11a934b75a
```
