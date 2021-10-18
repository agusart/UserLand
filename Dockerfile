FROM golang:latest

LABEL maintainer="Agus Budianto <agus.kbk@gmail.com>"

WORKDIR /app

COPY . .

RUN mkdir /asset

RUN go mod download

RUN go get -d github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build main.go" --command=./main