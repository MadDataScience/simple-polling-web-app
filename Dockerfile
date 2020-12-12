# https://devopsheaven.com/sqlite/backup/restore/dump/databases/docker/2017/10/10/sqlite-backup-restore-docker.html
# https://golangdocs.com/golang-docker
# https://medium.com/@petomalina/using-go-mod-download-to-speed-up-golang-docker-builds-707591336888

FROM golang:alpine

RUN apk add --no-cache git
RUN apk add --no-cache sqlite-libs sqlite-dev
RUN apk add --no-cache build-base
# RUN go get github.com/mattn/go-sqlite3

RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .  

# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o server .
RUN go build -o server . 

# FROM scratch
# COPY --from=build-env /app/server /app/server
# ENTRYPOINT ["/app/server"]

CMD [ "/app/server" ]
