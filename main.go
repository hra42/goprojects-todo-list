package main

import (
	"log"

	"github.com/hra42/goprojects-todo-list/internal/cli"
)

func main() {
	if err := cli.Start(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
