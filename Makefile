OS := $(shell uname)

build: */*.go
	go build 

all: build
	./gotask start

docker-build:
	docker build -t edwinlll/gotask:latest .

docker-push:
	docker push edwinlll/gotask:latest

docker-github:
	docker tag edwinlll/gotask:latest docker.pkg.github.com/lvzhihao/gotask/gotask:latest
	docker push docker.pkg.github.com/lvzhihao/gotask/gotask:latest

docker-ccr:
	docker tag edwinlll/gotask:latest ccr.ccs.tencentyun.com/wdwd/gotask:latest
	docker push ccr.ccs.tencentyun.com/wdwd/gotask:latest
	docker rmi ccr.ccs.tencentyun.com/wdwd/gotask:latest

docker-uhub:
	docker tag edwinlll/gotask:latest uhub.service.ucloud.cn/mmzs/gotask:latest
	docker push uhub.service.ucloud.cn/mmzs/gotask:latest
	docker rmi uhub.service.ucloud.cn/mmzs/gotask:latest
