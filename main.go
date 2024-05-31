package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/services/oidcrp"
	"github.com/datasektionen/logout/services/passkey"
	"github.com/datasektionen/logout/services/user"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	db, err := sqlx.Connect("pgx", config.Config.DatabaseURL.String())
	if err != nil {
		panic(err)
	}

	user := must(user.NewService(db))
	passkey := must(passkey.NewService(db))
	oidcrp := must(oidcrp.NewService(context.Background()))

	user.Assign(passkey)
	passkey.Assign(user)
	oidcrp.Assign(user)

	colonPort := ":" + strconv.Itoa(config.Config.Port)
	l, err := net.Listen("tcp", colonPort)
	if err != nil {
		slog.Error("Could not start listening for connections", "port", colonPort, "error", err)
		os.Exit(1)
	}
	slog.Info("Server started", "address", "http://localhost"+colonPort)
	slog.Error("Failed serving http server", "error", http.Serve(l, nil))
	os.Exit(1)
}
