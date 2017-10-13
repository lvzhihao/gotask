OS := $(shell uname)

build: */*.go
	go build 

all: build
	./gotask start

docker-build:
	sudo docker build -t edwinlll/gotask:latest .

docker-push:
	sudo docker push edwinlll/gotask:latest

docker-ccr:
	sudo docker tag edwinlll/gotask:latest ccr.ccs.tencentyun.com/wdwd/gotask
	sudo docker push ccr.ccs.tencentyun.com/wdwd/gotask
