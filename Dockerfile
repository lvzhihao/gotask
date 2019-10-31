FROM golang:1.13 as builder
ENV GOPROXY https://goproxy.io
WORKDIR /go/src/github.com/lvzhihao/gotask
COPY . . 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /usr/local/gotask
COPY --from=builder /go/src/github.com/lvzhihao/gotask/gotask .
ENV PATH /usr/local/gotask:$PATH
CMD ["gotask", "start"]
EXPOSE 8179
