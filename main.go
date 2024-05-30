package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
)

//go:generate templ generate

func main() {
	s, err := NewService()
	if err != nil {
		slog.Error("Could set shit up", "error", err)
		os.Exit(1)
	}

	http.Handle("GET /{$}", route(s.Index))
	http.Handle("GET /logout", route(s.Logout))
	http.Handle("GET /account", route(s.Account))
	http.Handle("GET /register", route(s.ShowRegister))
	http.Handle("POST /register", route(s.DoRegister))
	http.Handle("GET /login/dev", route(s.ShowLoginDev))
	http.Handle("POST /login/dev", route(s.DoLoginDev))
	http.Handle("GET /login/passkey", route(s.BeginLoginPasskey))
	http.Handle("POST /login/passkey", route(s.FinishLoginPasskey))
	http.Handle("GET /passkey/add", route(s.BeginAddPasskey))
	http.Handle("POST /passkey/add", route(s.FinishAddPasskey))
	http.Handle("POST /passkey/remove", route(s.RemovePasskey))

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("Could not start listening for connections", "port", port, "error", err)
		os.Exit(1)
	}
	slog.Info("Server started", "address", "http://localhost:"+port)
	slog.Error("Failed serving http server", "error", http.Serve(l, nil))
	os.Exit(1)
}
