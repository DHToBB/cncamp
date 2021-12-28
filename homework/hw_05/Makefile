export tag=v4.0

build:
	echo "building httpserver binary with metrics"
	mkdir -p bin/
	cd httpserver && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../bin/httpserver main.go

release:build
	echo "building httpserver container with metrics"
	docker build -t dhtobb/httpserver:${tag} .

push:release
	echo "pushing dhtobb/httpserver with metrics"
	docker push dhtobb/httpserver:${tag}

#部署一套完整的httpserver集群
deployment:
	# 创建名称空间
	kubectl create ns dhtobb-hs
	kubectl label ns dhtobb-hs istio-injection=enabled

	# 创建deployment
	kubectl apply -f specs/configmap.yaml -n dhtobb-hs

	kubectl apply -f specs/pv.yaml -n dhtobb-hs
	kubectl apply -f specs/pvc.yaml -n dhtobb-hs

	kubectl apply -f specs/deployment.yaml -n dhtobb-hs

	# 创建service
	kubectl apply -f specs/service.yaml -n dhtobb-hs

	# 创建secret, 注意 secret的命名空间是 istio-system，而不是dhtobb-hs
	#openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=dhtobb Inc./CN=*.dhtobb.io' -keyout dhtobb.io.key -out dhtobb.io.crt
	#kubectl create secret tls dhtobb-credential --key=dhtobb.io.key --cert=dhtobb.io.crt -n istio-system
	kubectl create -f specs/secret.yaml  -n istio-system

	# 创建istio
	kubectl apply -f specs/istio-specs.yaml -n dhtobb-hs


#	kubectl create configmap envoy-config --from-file=specs/envoy.yaml  -n dhtobb-hs
#	kubectl create -f specs/envoy-deploy.yaml -n dhtobb-hs



#测试最终结果
result:
	sleep 3
	curl --noproxy "*" -H "Host: dhtobb.com" https://$(shell kubectl get svc -n ingress-nginx | grep NodePort | awk -F' ' '{print $$3}') -v -k

clean:
	kubectl delete --ignore-not-found=true -f specs/configmap.yaml -n dhtobb-hs
	kubectl delete --ignore-not-found=true -f specs/deployment.yaml -n dhtobb-hs
	kubectl delete --ignore-not-found=true -f specs/service.yaml -n dhtobb-hs
	kubectl delete --ignore-not-found=true -f specs/pvc.yaml -n dhtobb-hs
	kubectl delete --ignore-not-found=true -f specs/pv.yaml -n dhtobb-hs

	kubectl delete --ignore-not-found=true -f specs/istio-specs.yaml -n dhtobb-hs
	kubectl delete --ignore-not-found=true -f specs/secret.yaml -n istio-system

	kubectl delete ns dhtobb-hs
#	kubectl delete --ignore-not-found=true -f specs/envoy-deploy.yaml
#	kubectl delete --ignore-not-found=true configmap envoy-config





