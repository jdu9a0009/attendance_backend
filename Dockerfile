FROM golang:1.21-alpine3.20

RUN mkdir /app

ADD . /app

WORKDIR /app

ENV TZ="Asia/Tashkent"

RUN go build -o main ./cmd/main.go

RUN go mod tidy && go mod vendor

CMD ["./main"]