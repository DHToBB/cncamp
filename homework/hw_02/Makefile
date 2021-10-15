export tag=v1.0

build:
	echo "building httpserver binary"
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/amd64 ./httpserver

release: build
	echo "building httpserver container"
	docker build -t dhtobb/httpserver:${tag} .

push: release
	echo "pushing dhtobb/httpserver"
	docker push dhtobb/httpserver:${tag}