package metrics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var logins = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "logout_logins",
		Help: "The total number of user logins",
	},
	[]string{"logout_login_type"},
)

var verifications = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "logout_legacyapi_verifications",
		Help: "The total number of legacyapi verifications",
	},
	[]string{"legacyapi_verification_type"},
)

func init() {
	prometheus.MustRegister(logins)
	prometheus.MustRegister(verifications)
}

func Start(ctx context.Context, port int) {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		<-ctx.Done()
		cc, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		slog.Info("shutting down metrics server")
		srv.Shutdown(cc)
		slog.Info("metrics server is shut down")
	}()

	slog.Info("starting metrics server", "port", port)

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start metrics server", "err", err)
	}
}

func IncLogin(loginType string) {
	logins.WithLabelValues(loginType).Inc()
}

func IncVerifications(status string) {
	verifications.WithLabelValues(status).Inc()
}
