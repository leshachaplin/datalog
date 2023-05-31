package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/leshachaplin/datalog/internal/domain"
)

type WorkerPool interface {
	Start(executeFn func(ctx context.Context, batch []domain.Event) error)
	GracefulStop()
	Process(payload domain.EventBatch)
	onFailure(payload domain.EventBatch, err error)
}

type Pool struct {
	numWorkers       int
	batchSize        int
	maxBatchCapacity int
	taskPayload      chan domain.EventBatch
	queue            Queue
	errorQueue       Queue
	start            sync.Once
	stop             sync.Once
	doneChan         chan struct{}
	ctx              context.Context //TODO: maybe make some wrapper func for getting context
	cancelFn         context.CancelFunc
	wg               *sync.WaitGroup
	logger           zerolog.Logger
}

func New(ctx context.Context, cfg Config, queue Queue, logger zerolog.Logger) *Pool {
	c, cancelFn := context.WithCancel(ctx)
	pool := Pool{
		numWorkers:  cfg.NumWorkers,
		taskPayload: make(chan domain.EventBatch, cfg.NumWorkers),
		doneChan:    make(chan struct{}),
		queue:       queue,
		ctx:         c,
		cancelFn:    cancelFn,
		wg:          &sync.WaitGroup{},
		logger:      logger,
	}
	if cfg.BatchSize <= 1 {
		pool.batchSize = 1000
	}
	if cfg.MaxBatchCapacity < 30 {
		pool.maxBatchCapacity = 30
	}
	return &pool
}

func (w *Pool) Start(
	executeFn func(ctx context.Context, eventBatch []domain.Event) error,
) {
	w.start.Do(func() {
		for i := 0; i < w.numWorkers; i++ {
			w.wg.Add(1)
			l := w.logger.With().Interface("worker", i).Logger()
			go w.work(w.ctx, l, executeFn)
		}

		go w.queue.Consume(w.ctx, w.taskPayload, w.doneChan)
	})
}

func (w *Pool) GracefulStop() {
	w.stop.Do(func() {
		close(w.doneChan)
		w.cancelFn()
		w.wg.Wait()
	})
}

func (w *Pool) Process(eventBatch domain.EventBatch) {
	if err := w.queue.Publish(w.ctx, uuid.NewString(), eventBatch); err != nil {
		w.onFailure(eventBatch, err)
	}
}

func (w *Pool) onFailure(eventBatch domain.EventBatch, err error) {
	p := payload{
		Payload: eventBatch,
	}
	p.SetErrorReason(err)
	if errPublish := w.errorQueue.Publish(w.ctx, eventBatch.ID, p); errPublish != nil {
		log.Err(err).Interface("EventBatch", eventBatch).Msg("failed to process events")
	}
}

// TODO: сделать так чтобы батчи собирались в фиксированный размер из конфига - минимум 1000
func (w *Pool) work(
	ctx context.Context,
	logger zerolog.Logger,
	executeFn func(ctx context.Context, eventBatch []domain.Event) error,
) {
	defer w.wg.Done()
	batch := make([]domain.Event, 0, w.batchSize)
	batchID := uuid.NewString()
	defer func() {
		if len(batch) > 0 {
			logger.Debug().Interface("EVENTS", batch).Msg("start processing events")
			if err := executeFn(context.Background(), batch); err != nil {
				w.onFailure(domain.EventBatch{
					ID:     batchID,
					Events: batch,
				}, err)
			}
			logger.Debug().Str("EVENTS", "end processing events").Send()
			batch = nil
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.doneChan:
			return
		case pld, ok := <-w.taskPayload:
			if !ok {
				return
			}

			batch = append(batch, pld.Events...)
			if len(batch) >= 1000 {
				logger.Debug().Str("BATCH_ID", pld.ID).Interface("EVENTS", pld.Events).Msg("start processing events")
				if err := executeFn(context.Background(), pld.Events); err != nil {
					w.onFailure(pld, err)
				}
				batchID = pld.ID
				logger.Debug().Str("BATCH_ID", pld.ID).Msg("end processing events")
				if w.maxBatchCapacity > cap(batch) {
					batch = nil
					batch = make([]domain.Event, 0, w.batchSize)
				} else {
					batch = batch[:0]
				}
			}
		}
	}
}

type payload struct {
	Payload domain.EventBatch `json:"payload"`
	Error   *errorReason      `json:"error_reason"`
}

func (c *payload) SetErrorReason(err error) {
	if c.Error == nil {
		c.Error = new(errorReason)
	}
	c.Error.Reason = err
}

func (c *payload) GetErrorReason() error {
	if c.Error != nil {
		return c.Error.Reason
	}
	return nil
}

type errorReason struct {
	Reason error
}

func (e errorReason) MarshalJSON() ([]byte, error) {
	if e.Reason != nil {
		return json.Marshal(e.Reason.Error())
	}
	return json.Marshal(nil)
}

func (e *errorReason) UnmarshalJSON(data []byte) error {
	e.Reason = errors.New(string(data))
	return nil
}
