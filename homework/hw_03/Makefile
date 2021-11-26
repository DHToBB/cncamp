export tag=v2.0

build:
	echo "building httpserver binary"
	mkdir -p bin/
	cd httpserver && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../bin/httpserver main.go

release:build
	echo "building httpserver container"
	docker build -t dhtobb/httpserver:${tag} .

push:release
	echo "pushing dhtobb/httpserver"
	docker push dhtobb/httpserver:${tag}