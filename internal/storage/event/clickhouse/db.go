package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	//"github.com/golang-migrate/migrate/v4"
	//ch "github.com/golang-migrate/migrate/v4/database/clickhouse"
)

type Clickhouse struct {
	conn driver.Conn
}

func New(ctx context.Context, cfg Config) (*Clickhouse, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.DB,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Debug: true,
		Debugf: func(format string, v ...any) {
			fmt.Printf(format, v)
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:     time.Second * 30,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Duration(10) * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	return &Clickhouse{
		conn: conn,
	}, nil
}

func (c *Clickhouse) Close() error {
	return c.conn.Close()
}

func (c *Clickhouse) Migrate(ctx context.Context) error {
	return c.conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS events
		(
    		client_time DATETIME,
    		server_time DATETIME,
    		ip          IPv4,
    		device_id   String,
    		device_os   String,
    		session     String,
    		sequence    Int16,
    		event_type  String,
    		param_int   Int32,
    		param_str   String
		) Engine = MergeTree
		ORDER BY server_time`)
}
