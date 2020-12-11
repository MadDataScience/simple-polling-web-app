package main

import (
	"github.com/maddatascience/simple-polling-web-app/cmd"
)

// var (
// 	BuildCommitHash string
// 	BuildTime       string
// )

func main() {
	// println("Build time:" + BuildTime + ", Build commit hash:" + BuildCommitHash)
	err := cmd.Execute()
	if err != nil {
		println(err)
	}
}
