package main

import (
	"context"
	"os"

	"github.com/pressly/goose/v3"
	_ "github.com/datasektionen/logout/pkg/database"
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: go run ./cmd/goose <command> <args...>")
	}
	goose.RunContext(context.Background(), os.Args[1], nil, "pkg/database/migrations", os.Args[2:]...)
}
