package main

import (
	"file-conv/internal/routes"
)

func main() {
	server := routes.NewApiServer(":3000")
	server.Run()
}
