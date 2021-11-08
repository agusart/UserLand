FROM golang:latest

LABEL maintainer="Agus Budianto <agus.kbk@gmail.com>"

RUN mkdir /asset

WORKDIR /app

COPY . .

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build userland" --command=./userland