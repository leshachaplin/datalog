package worker

import (
	"context"

	"github.com/leshachaplin/datalog/internal/domain"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/consumer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/producer"
)

type Queue interface {
	Publish(ctx context.Context, key string, payload any) error
	Consume(ctx context.Context, taskPayload chan<- domain.EventBatch, done <-chan struct{})
}

type RedpandaQueue struct {
	producer *producer.Producer
	consumer *consumer.Consumer
}

func NewRedpandaQueue(producer *producer.Producer, consumer *consumer.Consumer) *RedpandaQueue {
	return &RedpandaQueue{
		producer: producer,
		consumer: consumer,
	}
}

func (r *RedpandaQueue) Publish(ctx context.Context, key string, payload any) error {
	if err := r.producer.Publish(ctx, key, payload); err != nil {
		return err
	}
	return nil
}

func (r *RedpandaQueue) Consume(ctx context.Context, taskPayload chan<- domain.EventBatch, done <-chan struct{}) {
	r.consumer.Consume(ctx, taskPayload, done)
}
