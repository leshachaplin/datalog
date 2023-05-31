package config

import (
	"github.com/leshachaplin/datalog/internal/storage/event/clickhouse"
	"github.com/leshachaplin/datalog/internal/worker"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/consumer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/producer"
)

// Config is the main config for the application
type Config struct {
	LogLevel      string            `mapstructure:"log_level"`
	Clickhouse    clickhouse.Config `mapstructure:"clickhouse"`
	EventWorker   worker.Config     `mapstructure:"auth_url"`
	EventProducer producer.Config   `mapstructure:"event_producer"`
	EventConsumer consumer.Config   `mapstructure:"event_consumer"`
}
