package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/services/dev"
	"github.com/datasektionen/logout/services/legacyapi"
	"github.com/datasektionen/logout/services/oidcrp"
	"github.com/datasektionen/logout/services/passkey"
	"github.com/datasektionen/logout/services/static"
	"github.com/datasektionen/logout/services/user"
	"github.com/datasektionen/logout/services/oidcprovider"
)

func main() {
	db, err := database.ConnectAndMigrate(context.Background())
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	user := must(user.NewService(db))
	passkey := must(passkey.NewService(db))
	oidcrp := must(oidcrp.NewService(ctx))
	legacyapi := must(legacyapi.NewService(ctx, db))
	dev := must(dev.NewService(db))
	oidcprovider := must(oidcprovider.NewService(db))
	cancel()

	user.Assign(dev)
	passkey.Assign(user)
	oidcrp.Assign(user)
	legacyapi.Assign(user)
	dev.Assign(user)
	oidcprovider.Assign(user)

	static.Mount()

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

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
