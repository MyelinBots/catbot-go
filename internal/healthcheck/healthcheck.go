package healthcheck

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/MyelinBots/catbot-go/config"
)

// Healthcheck that starts http server
func StartHealthcheck(ctx context.Context, cfg config.AppConfig) {
	// start http server
	go func() {
		port := strconv.Itoa(cfg.Port)
		err := http.ListenAndServe(":"+port, HealthCheckHandler())
		if err != nil && err != http.ErrServerClosed {
			log.Printf("healthcheck server error: %v", err)
		}
	}()

}

func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
