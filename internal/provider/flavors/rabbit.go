package flavors

import (
	"context"
	"fmt"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitDriver struct {
	logger *logz.LoggerZ
	conn   *amqp091.Connection
}

func NewRabbitDriver(logger *logz.LoggerZ) types.Driver {
	return &RabbitDriver{
		logger: logger,
	}
}

func (d *RabbitDriver) Connect(ctx context.Context, cfg *types.DBConfig) error {
	if cfg == nil {
		return fmt.Errorf("rabbitmq: nil config")
	}
	if cfg.DSN == "" {
		return fmt.Errorf("rabbitmq: empty DSN")
	}

	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := amqp091.DialConfig(cfg.DSN, amqp091.Config{
		Dial: amqp091.DefaultDial(10 * time.Second),
	})
	if err != nil {
		return fmt.Errorf("rabbitmq: dial error: %w", err)
	}

	d.conn = conn
	return nil
}

func (d *RabbitDriver) Ping(ctx context.Context) bool {
	if d.conn == nil {
		return false
	}

	// RabbitMQ não tem ping real — usamos Channel open/close como heartbeat.
	ch, err := d.conn.Channel()
	if err != nil {
		return false
	}
	_ = ch.Close()
	return true
}

func (d *RabbitDriver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}
