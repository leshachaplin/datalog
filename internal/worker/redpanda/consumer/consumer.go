package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/leshachaplin/datalog/internal/domain"
)

const (
	defaultPollFetchesTimeout = 15 * time.Second
	defaultRetryCount         = 10
)

type Config struct {
	Brokers            []string      `mapstructure:"brokers"`
	ConsumerGroup      string        `mapstructure:"consumer_group"`
	Topics             []string      `mapstructure:"topics"`
	RetryCount         int           `mapstructure:"retry_count"`
	PollFetchesTimeout time.Duration `mapstructure:"poll_fetches_timeout"`
}

type Consumer struct {
	client             *kgo.Client
	retryCount         int
	pollFetchesTimeout time.Duration
	errChan            chan<- error
}

func NewConsumer(cfg Config, errChan chan<- error) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(cfg.Topics...),
		kgo.DisableAutoCommit(),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("kgo new client: %w", err)
	}

	ctx, cansel := context.WithTimeout(context.Background(), time.Second*15)
	defer cansel()
	if err := client.Ping(ctx); err != nil {
		return nil, err
	}

	consumer := &Consumer{
		client:  client,
		errChan: errChan,
	}

	if cfg.PollFetchesTimeout == 0 {
		consumer.pollFetchesTimeout = defaultPollFetchesTimeout
	} else {
		consumer.pollFetchesTimeout = cfg.PollFetchesTimeout
	}

	if cfg.RetryCount == 0 {
		consumer.retryCount = defaultRetryCount
	} else {
		consumer.retryCount = cfg.RetryCount
	}

	return consumer, nil
}

func (c *Consumer) Close() error {
	c.client.Close()
	return nil
}

func (c *Consumer) Consume(ctx context.Context, eventChan chan<- domain.EventBatch, done <-chan struct{}) {
	c.consume(ctx, done, func(fetches kgo.Fetches) error {
		for iter := fetches.RecordIter(); !iter.Done(); {
			record := iter.Next()

			var event domain.EventBatch
			if err := json.Unmarshal(record.Value, &event); err != nil {
				log.Error().Str("record", string(record.Value)).Err(err).Msg("Consume: Unmarshal event value.")

				if commitErr := c.client.CommitRecords(ctx, record); commitErr != nil {
					return fmt.Errorf("commit record: %w", commitErr)
				}
				return err
			}
			eventChan <- event
			if commitErr := c.client.CommitRecords(ctx, record); commitErr != nil {
				return fmt.Errorf("commit record: %w", commitErr)
			}
		}
		return nil
	})
}

func (c *Consumer) consume(ctx context.Context, done <-chan struct{}, fn func(fetches kgo.Fetches) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		default:
			fetchCtx, cancel := context.WithTimeout(ctx, c.pollFetchesTimeout)
			fetches := c.client.PollFetches(fetchCtx)
			cancel()

			if fetches.IsClientClosed() {
				c.errChan <- errors.New("client closed")
				return
			}

			if err := fetches.Err(); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}

				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}

				c.errChan <- fmt.Errorf("stream poll fetches: %w", err)
				continue
			}

			if err := fn(fetches); err != nil {
				continue
			}
		}
	}
}
