package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/datasektionen/logout/database"
	"github.com/datasektionen/logout/handlers"
	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/oidcprovider"
	"github.com/datasektionen/logout/pkg/static"
	"github.com/datasektionen/logout/service"
)

func main() {
	initCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	db, err := database.ConnectAndMigrate(initCtx)
	if err != nil {
		panic(err)
	}

	s, err := service.NewService(initCtx, db)
	if err != nil {
		panic(err)
	}
	if err := oidcprovider.MountRoutes(s); err != nil {
		panic(err)
	}
	cancel()

	handlers.MountRoutes(s)

	static.Mount()

	port := strconv.Itoa(config.Config.Port)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Could not start listening for connections", "port", port, "error", err)
		os.Exit(1)
	}
	slog.Info("Server started", "address", "http://localhost:"+port)
	slog.Error("Failed serving http server", "error", http.Serve(l, nil))
	os.Exit(1)
}
