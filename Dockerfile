FROM golang:latest

LABEL maintainer="Agus Budianto <agus.kbk@gmail.com>"

WORKDIR /app

COPY . .

RUN mkdir /asset

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build userland" --command=./userland