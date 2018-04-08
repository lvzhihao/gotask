FROM golang:1.9 as builder
WORKDIR /go/src/github.com/lvzhihao/gotask
COPY . . 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /usr/local/gotask
COPY --from=builder /go/src/github.com/lvzhihao/gotask/gotask .
COPY ./docker-entrypoint.sh  .
ENV PATH /usr/local/gotask:$PATH
RUN chmod +x /usr/local/gotask/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/gotask/docker-entrypoint.sh"]
EXPOSE 8179
