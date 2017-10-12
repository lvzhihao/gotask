OS := $(shell uname)

build: */*.go
	go build 

all: build
	./gotask start

docker-build:
	sudo docker build -t edwinlll/gotask:latest .

docker-push:
	sudo docker push edwinlll/gotask:latest
