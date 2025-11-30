package flavors

import (
	"context"
	"fmt"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDriver struct {
	logger *logz.LoggerZ
	client *mongo.Client
}

func NewMongoDriver(logger *logz.LoggerZ) types.Driver {
	return &MongoDriver{
		logger: logger,
	}
}

func (d *MongoDriver) Connect(ctx context.Context, cfg *types.DBConfig) error {
	if cfg == nil {
		return fmt.Errorf("mongodb: nil config")
	}
	if cfg.DSN == "" {
		return fmt.Errorf("mongodb: empty DSN")
	}

	clientOpts := options.Client().ApplyURI(cfg.DSN)

	ctxConn, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctxConn, clientOpts)
	if err != nil {
		return fmt.Errorf("mongodb: connect error: %w", err)
	}

	ctxPing, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()

	if err := client.Ping(ctxPing, nil); err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("mongodb: ping fail: %w", err)
	}

	d.client = client
	return nil
}

func (d *MongoDriver) Ping(ctx context.Context) bool {
	if d.client == nil {
		return false
	}

	ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return d.client.Ping(ctxPing, nil) == nil
}

func (d *MongoDriver) Close() error {
	if d.client != nil {
		return d.client.Disconnect(context.Background())
	}
	return nil
}
