FROM golang:1.9

COPY . /go/src/github.com/lvzhihao/gotask 

WORKDIR /go/src/github.com/lvzhihao/gotask

RUN go-wrapper install

CMD ["go-wrapper", "run", "start"]

EXPOSE 8179
