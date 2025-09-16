package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/handlers"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/oidcprovider"
	"github.com/datasektionen/sso/pkg/static"
	"github.com/datasektionen/sso/service"
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
	op, err := oidcprovider.Init(s)
	if err != nil {
		panic(err)
	}
	cancel()

	if config.Config.PortExternal != 0 {
		go serve(s, op, false, config.Config.PortExternal)
	}
	serve(s, op, true, config.Config.PortInternal)
}

func serve(s *service.Service, op http.Handler, internal bool, p int) {
	mux := http.NewServeMux()
	handlers.MountRoutes(s, mux, internal)
	mux.Handle("/op/", op)
	static.Mount(mux)

	port := strconv.Itoa(p)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Could not start listening for connections", "port", port, "error", err)
		os.Exit(1)
	}
	started := "External server started"
	if internal {
		started = "Internal server started"
	}
	slog.Info(started, "address", "http://localhost:"+port, "config", config.Config)
	slog.Error("Failed serving http server", "error", http.Serve(l, mux))
	os.Exit(1)
}
