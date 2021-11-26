##课后练习5.1：

---

- 在本地构建一个单节点的基于HTTPS的etcd集群
- 写一条数据
- 查看数据细节
- 删除数据

##docker安装
```shell
#拉取镜像
docker pull registry.aliyuncs.com/google_containers/etcd:3.5.0-0

#启动容器
docker run -d registry.aliyuncs.com/google_containers/etcd:3.5.0-0 /usr/local/bin/etcd

#进入容器
docker exec -it <containerid> sh

#写入数据
etcdctl put x 0

#读取数据
sh-5.0# etcdctl get x -w=json
{"header":{"cluster_id":14841639068965178418,"member_id":10276657743932975437,"revision":2,"raft_term":2},"kvs":[{"key":"eA==","create_revision":2,"mod_revision":2,"version":1,"value":"MA=="}],"count":1}

#删除数据
sh-5.0# etcdctl del x
1

```

##二进制安装

---
1. 下载cfssl工具
> https://github.com/cloudflare/cfssl/releases

也可以使用wget下载（如果能够连接上网）
```shell
wget https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
wget https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
wget https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64
```

添加执行权限，并拷贝到可执行路径
```shell
chmod +x cfssl*
mv cfssl_linux-amd64 /usr/local/bin/cfssl
mv cfssljson_linux-amd64 /usr/local/bin/cfssljson
mv cfssl-certinfo_linux-amd64 /usr/local/bin/cfssl-certinfo
```

2. 生成CA证书配置文件
```shell
cat > ca-csr.json <<"EOF"
{
"CN": "kubernetes",
"key": {
"algo": "rsa",
"size": 2048
},
"names": [
{
"C": "CN",
"ST": "Shanghai",
"L": "Shanghai",
"O": "cncamp",
"OU": "cncamp"
}
],
"ca": {
"expiry": "87600h"
}
}
EOF
```

3. 生成ca证书文件
```shell
cfssl gencert -initca ca-csr.json | cfssljson -bare ca

#生成三个文件： ca.csr  ca-key.pem  ca.pem
```

4. 配置CA证书策略
```shell
cat > ca-config.json <<"EOF"
{
  "signing": {
      "default": {
          "expiry": "87600h"
        },
      "profiles": {
          "kubernetes": {
              "usages": [
                  "signing",
                  "key encipherment",
                  "server auth",
                  "client auth"
              ],
              "expiry": "87600h"
          }
      }
  }
}
EOF
```

5. 生成etcd请求的csr文件
```shell
cat > etcd-csr.json <<"EOF"
{
  "CN": "etcd",
  "hosts": [
    "127.0.0.1",
    "192.168.34.2"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [{
    "C": "CN",
    "ST": "Shanghai",
    "L": "Shanghai",
    "O": "cncamp",
    "OU": "cncamp"
  }]
}
EOF
```

6. 生成证书
```shell
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes etcd-csr.json | cfssljson  -bare etcd

#生成文件：  etcd.csr  etcd-key.pem  etcd.pem
``` 

7. 下载etcd源文件
> https://github.com/etcd-io/etcd/tags

```shell
tar -xvzf etcd-v3.5.1-linux-amd64.tar.gz
cp -p etcd-v3.5.1-linux-amd64/etcd* /usr/local/bin/
```

8. 生成etcd配置文件
```shell
cat >  etcd.conf <<"EOF"
#[Member]
ETCD_NAME="etcd1"
ETCD_DATA_DIR="/var/lib/etcd/default.etcd"
ETCD_LISTEN_PEER_URLS="https://192.168.34.2:2380"
ETCD_LISTEN_CLIENT_URLS="https://192.168.34.2:2379,http://127.0.0.1:2379"

#[Clustering]
ETCD_INITIAL_ADVERTISE_PEER_URLS="https://192.168.34.2:2380"
ETCD_ADVERTISE_CLIENT_URLS="https://192.168.34.2:2379"
ETCD_INITIAL_CLUSTER="etcd1=https://192.168.34.2:2380"
ETCD_INITIAL_CLUSTER_TOKEN="etcd-cluster"
ETCD_INITIAL_CLUSTER_STATE="new"
EOF
```

9. 启动etcd
```shell
```

