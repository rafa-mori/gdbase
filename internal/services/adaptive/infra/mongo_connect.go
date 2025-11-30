package infra

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// BuildMongoURI monta a URI de conexão. Se appUser/appPass vierem vazios,
// usa cfg.Username/cfg.Password. authSource default = "admin".
func BuildMongoURI(cfg *types.DBConfig) (string, string) {
	user := cfg.User
	pass := cfg.Pass
	host := cfg.Host
	port := cfg.Port
	db := cfg.DBName
	if db == "" {
		db = "canalizeds"
	}
	usr := url.QueryEscape(user)
	pwd := url.QueryEscape(pass)
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin", usr, pwd, host, port, db)
	return uri, db
}

// ConnectMongo abre pool de conexão, faz ping e retorna client+db.
// Se appUser/appPass vazios, usa credenciais do cfg.
func ConnectMongo(ctx context.Context, cfg *types.DBConfig) (*mongo.Client, *mongo.Database, error) {
	if !kbx.DefaultTrue(cfg.Enabled) {
		return nil, nil, fmt.Errorf("mongo disabled or nil config")
	}
	uri, dbName := BuildMongoURI(cfg)

	clientOpts := options.Client().ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(8 * time.Second).
		SetRetryWrites(true).
		SetMaxPoolSize(40)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}

	// Ping para garantir readiness real do driver
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}

	logz.Log("info", "Mongo connected", "db", dbName, "host", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

	return client, client.Database(dbName), nil
}

// CloseMongo encerra o pool com timeout.
func CloseMongo(ctx context.Context, client *mongo.Client) {
	if client == nil {
		return
	}
	ctxClose, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Disconnect(ctxClose); err != nil {
		logz.Log("error", "Mongo disconnect error", "err", err)
	}
}

// PingMongo expõe um ping simples para healthchecks.
func PingMongo(ctx context.Context, client *mongo.Client) error {
	if client == nil {
		return fmt.Errorf("nil client")
	}
	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return client.Ping(ctxPing, readpref.Primary())
}
