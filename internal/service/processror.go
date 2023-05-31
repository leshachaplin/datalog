package service

import (
	"context"

	"github.com/leshachaplin/datalog/internal/domain"
	"github.com/leshachaplin/datalog/internal/worker"
)

type Storage interface {
	StoreEvents(ctx context.Context, events []domain.Event) error
}

type Processor interface {
	Event
	Storage
}

type Service struct {
	eventPool    worker.WorkerPool
	eventStorage Storage
}

func New(eventPool worker.WorkerPool, eventStorage Storage) *Service {
	eventPool.Start(eventStorage.StoreEvents)

	return &Service{
		eventPool:    eventPool,
		eventStorage: eventStorage,
	}
}
