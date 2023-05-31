package producer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Config struct {
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
	SleepDuration time.Duration `mapstructure:"sleep_duration"`
	Brokers       []string      `mapstructure:"brokers"`
	Topic         string        `mapstructure:"topic"`
}

type Producer struct {
	retryAttempts int
	retryDelay    time.Duration
	sleepDuration time.Duration
	client        *kgo.Client
	logger        zerolog.Logger
}

func NewProducer(
	ctx context.Context,
	cfg Config,
	logger zerolog.Logger,
) (*Producer, error) {
	clientOpts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.Topic),
	}

	client, err := kgo.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("kgo new client: %w", err)
	}

	if err = client.Ping(ctx); err != nil {
		return nil, err
	}

	producer := &Producer{
		client:        client,
		retryAttempts: cfg.RetryAttempts,
		retryDelay:    cfg.RetryDelay,
		logger:        logger,
	}

	return producer, nil
}

func (p *Producer) Close() error {
	p.client.Close()
	return nil
}

func (p *Producer) Publish(ctx context.Context, key string, msg any) error {
	const publishTimeout = 5 * time.Second

	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	record := kgo.KeyStringRecord(key, string(b))

	return linearBackOff(&p.logger, p.retryAttempts, p.retryDelay, func() error {
		produceCtx, cancel := context.WithTimeout(ctx, publishTimeout)
		res := p.client.ProduceSync(produceCtx, record)
		cancel()

		if err := res.FirstErr(); err != nil {
			return fmt.Errorf("produce sync: %w", err)
		}
		return nil
	})
}

func linearBackOff(log *zerolog.Logger, attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
		} else {
			return nil
		}

		log.Warn().Err(err).Msgf("Retry: %d.", i)

		time.Sleep(delay * time.Duration(i+1))
	}
	return err
}
