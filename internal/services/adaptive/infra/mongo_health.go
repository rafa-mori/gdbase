package infra

import (
	"context"
	"net/http"
	"time"

	logz "github.com/kubex-ecosystem/logz"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoHealthHandler struct {
	Client *mongo.Client
}

func (h *MongoHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := PingMongo(ctx, h.Client); err != nil {
		logz.Log("error", "mongo unhealthy", "err", err)
		http.Error(w, "mongo unhealthy", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
