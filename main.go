package main

import (
	"os"

	"github.com/maddatascience/simple-polling-web-app/cmd"
)

const defaultPort = "8080"
const defaultDataSource = "poll.db"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	dataSourceName := os.Getenv("DB")
	if dataSourceName == "" {
		dataSourceName = defaultDataSource
	}
	err := cmd.Execute(port, dataSourceName)
	if err != nil {
		println(err)
	}
}
