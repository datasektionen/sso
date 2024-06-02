package main

import (
	"context"
	"os"

	"github.com/datasektionen/logout/pkg/database"
)

func main() {
	args := os.Args[1:]
	shift := func() string {
		if len(args) == 0 {
			panic("Expected another argument")
		}
		s := args[0]
		args = args[1:]
		return s
	}

	ctx := context.Background()
	switch shift() {
	case "add-user":
		db := must(database.Connect(ctx))
		assert(db.CreateUser(ctx, shift()))
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
