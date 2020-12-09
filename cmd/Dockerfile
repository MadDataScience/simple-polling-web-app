# https://devopsheaven.com/sqlite/backup/restore/dump/databases/docker/2017/10/10/sqlite-backup-restore-docker.html
# https://golangdocs.com/golang-docker

FROM golang:alpine

RUN apk add --no-cache git
RUN apk add --no-cache sqlite-libs sqlite-dev
RUN apk add --no-cache build-base
RUN go get github.com/mattn/go-sqlite3
RUN go get golang.org/x/crypto/bcrypt

RUN mkdir /app
COPY . /app
WORKDIR /app

RUN go build -o server . 

CMD [ "/app/server" ]
