export tag=v3.0

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
deployment:release
	# 创建deployment
	kubectl apply -f specs/configmap.yaml

	kubectl apply -f specs/pv.yaml
	kubectl apply -f specs/pvc.yaml

	kubectl apply -f specs/deployment.yaml

	# 创建service
	kubectl apply -f specs/service.yaml

	# 创建ingress
	## install ingress controller
	kubectl create -f specs/nginx-ingress-deployment.yaml

	# 创建secret
	kubectl create -f specs/secret.yaml

	# 创建 ingress
	kubectl delete -A --ignore-not-found=true validatingwebhookconfigurations ingress-nginx-admission
	kubectl create -f specs/ingress.yaml

	# 查看ingress启动情况
	kubectl get svc -n ingress-nginx

#测试最终结果
result:
	sleep 3
	curl --noproxy "*" -H "Host: dhtobb.com" https://$(shell kubectl get svc -n ingress-nginx | grep NodePort | awk -F' ' '{print $$3}') -v -k

clean:
	kubectl delete --ignore-not-found=true -f specs/configmap.yaml
	kubectl delete --ignore-not-found=true -f specs/deployment.yaml
	kubectl delete --ignore-not-found=true -f specs/service.yaml
	kubectl delete --ignore-not-found=true -f specs/secret.yaml
	kubectl delete -A --ignore-not-found=true validatingwebhookconfigurations ingress-nginx-admission
	kubectl delete --ignore-not-found=true --force=true --cascade=orphan -f specs/nginx-ingress-deployment.yaml #耗时较长
	kubectl delete --ignore-not-found=true -f specs/ingress.yaml
	kubectl delete --ignore-not-found=true -f specs/pvc.yaml
	kubectl delete --ignore-not-found=true -f specs/pv.yaml





