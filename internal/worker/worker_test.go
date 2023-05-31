package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/leshachaplin/datalog/internal/domain"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/consumer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/producer"
)

func (i *IntegrationTestSuite) TestWorker_RedpandaQueue() {
	cases := map[string]struct {
		cfg        Config
		taskAmount int
	}{
		"ok": {
			cfg:        Config{NumWorkers: 100},
			taskAmount: 100,
		},
		"ok - tasks more than workers": {
			cfg:        Config{NumWorkers: 10},
			taskAmount: 1000,
		},
		"ok - tasks less than workers": {
			cfg:        Config{NumWorkers: 200},
			taskAmount: 100,
		},
	}

	for name, tc := range cases {
		i.Run(name, func() {
			defer goleak.VerifyNone(i.T())
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)

			consumerErrorChan := make(chan error, 1)
			defer close(consumerErrorChan)
			consumer, err := consumer.NewConsumer(i.consumerCfg, consumerErrorChan)
			i.Require().NoError(err)
			defer consumer.Close()

			producer, err := producer.NewProducer(
				i.ctx,
				i.producerCfg,
				log.With().Str("producer", "Publish").Logger(),
			)
			i.Require().NoError(err)

			payloadChan := make(chan domain.EventBatch, tc.cfg.NumWorkers)
			execFn := func(ctx context.Context, payload domain.EventBatch) error {
				payloadChan <- payload
				return nil
			}

			l := log.With().Str("WORKER", "PROCESS").Logger()
			worker := New(ctx, tc.cfg, NewRedpandaQueue(producer, consumer), l)
			worker.Start(execFn)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func(t *testing.T) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case payload, ok := <-payloadChan:
						require.True(t, ok)
						require.Equal(t, "test_id", payload.ID)
					}
				}
			}(i.T())

			for k := 0; k < tc.taskAmount; k++ {
				event := domain.EventBatch{
					ID: "test_id",
					Events: []domain.Event{
						{
							DeviceID: uuid.New().String(),
						},
					},
				}
				worker.Process(event)
			}

			worker.GracefulStop()
			err = consumer.Close()
			i.NoError(err)
			err = producer.Close()
			i.NoError(err)
			cancel()
			wg.Wait()
		})
	}
}
