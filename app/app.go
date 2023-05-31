package app

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/leshachaplin/datalog/app/waiter"
	"github.com/leshachaplin/datalog/internal/config"
	appServer "github.com/leshachaplin/datalog/internal/server/http"
	"github.com/leshachaplin/datalog/internal/service"
	"github.com/leshachaplin/datalog/internal/storage/event/clickhouse"
	"github.com/leshachaplin/datalog/internal/worker"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/consumer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/producer"
)

const (
	defaultAddr = ":8080"
)

type LoadConfigFn func() (config.Config, error)

type App struct {
	cfg      config.Config
	logger   zerolog.Logger
	server   *appServer.Server
	waiter   waiter.Waiter
	ctx      context.Context
	cancelFn context.CancelFunc
}

func New(loadConfigFn LoadConfigFn) *App {
	ctx, cancelFn := context.WithCancel(context.Background())
	cfg, err := loadConfigFn()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	logger := NewZeroLogger(Level(cfg.LogLevel))

	w := waiter.NewWaiter(ctx, cancelFn)

	return &App{
		cfg:      cfg,
		logger:   logger,
		waiter:   w,
		ctx:      ctx,
		cancelFn: cancelFn,
	}
}

func (a *App) Start() {
	defer a.cancelFn()

	consumerErrorChan := make(chan error, 1)
	defer close(consumerErrorChan)
	eventConsumer, err := consumer.NewConsumer(a.cfg.EventConsumer, consumerErrorChan)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("Could not setup event consumer.")
	}
	defer eventConsumer.Close()

	eventProducer, err := producer.NewProducer(
		a.ctx,
		a.cfg.EventProducer,
		a.logger.With().Str("event producer", "Publish").Logger(),
	)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("Could not setup event producer.")
	}
	defer eventProducer.Close()

	eventQueue := worker.NewRedpandaQueue(eventProducer, eventConsumer)
	l := a.logger.With().Str("WORKER", "EVENT").Logger()
	eventWorker := worker.New(a.ctx, a.cfg.EventWorker, eventQueue, l)

	eventStorage, err := clickhouse.New(a.ctx, a.cfg.Clickhouse)
	if err != nil {
		a.logger.Fatal().Err(err).Msg("Could not setup event storage.")
	}
	defer eventStorage.Close()

	eventProcessor := service.New(eventWorker, eventStorage)
	handler := appServer.NewHandler(eventProcessor, a.logger)

	a.server = appServer.New(handler)

	a.waitForServer()
	a.waitForWorker(eventWorker)

	if err = a.waiter.Wait(); err != nil {
		a.logger.Fatal().Err(err).Msg("App crash.")
	}
}

func (a *App) Stop() {
	a.cancelFn()
}

func (a *App) waitForServer() {
	a.waiter.Add(func(ctx context.Context) error {
		defer a.logger.Debug().Msg("server has been shutdown")

		group, gCtx := errgroup.WithContext(ctx)
		group.Go(func() error {
			defer a.logger.Debug().Msg("public server exited")
			a.logger.Info().Str("starting server at: ", defaultAddr).Send()
			err := a.server.ServePublic(defaultAddr)
			if err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		})

		group.Go(func() error {
			<-gCtx.Done()
			log.Debug().Msg("shutting down the server")
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			if err := a.server.ShutdownPublic(ctx); err != nil {
				a.logger.Warn().Err(err).Msg("error while shutting down the server")
			}
			return nil
		})

		return group.Wait()
	})
}

func (a *App) waitForWorker(eventWorker worker.WorkerPool) {
	a.waiter.Add(func(ctx context.Context) error {
		group, gCtx := errgroup.WithContext(ctx)
		group.Go(func() error {
			<-gCtx.Done()
			eventWorker.GracefulStop()
			return nil
		})
		return group.Wait()
	})
}
