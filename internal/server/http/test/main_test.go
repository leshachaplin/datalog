package http

import (
	"context"
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/leshachaplin/datalog/app"
	"github.com/leshachaplin/datalog/internal/config"
	"github.com/leshachaplin/datalog/internal/storage/event/clickhouse"
	"github.com/leshachaplin/datalog/internal/worker"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/consumer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/producer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/testingh"
)

const (
	eventTopic        = "event"
	defaultAddrPublic = ":8080"
)

var (
	defaultTopics = []string{eventTopic}
)

type IntegrationTestSuite struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	kafkaCLi            *kadm.Client
	redpandaContainer   *testingh.Container
	clickhouseContainer *clickhouse.Container
	broker              string
	app                 *app.App

	cfg *config.Config
	*rand.Rand

	wg *sync.WaitGroup

	suite.Suite
}

func (i *IntegrationTestSuite) SetupSuite() {
	var err error
	ctx, cnsl := context.WithTimeout(context.Background(), time.Minute*2)
	i.ctx = ctx
	i.cancelFn = cnsl
	i.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	i.redpandaContainer, err = testingh.NewContainer(func(connURL string) error {
		i.broker = connURL
		opts := []kgo.Opt{
			kgo.SeedBrokers(connURL),
		}

		pandaCLi, err := kgo.NewClient(opts...)
		if err != nil {
			return err
		}

		pingErr := pandaCLi.Ping(ctx)
		if pingErr != nil {
			pandaCLi.Close()
			return pingErr
		}

		i.kafkaCLi = kadm.NewClient(pandaCLi)
		return nil
	})
	i.Require().NoError(err)

	var clickhouseAddr string
	i.clickhouseContainer, err = clickhouse.NewContainer(func(connURL string) error {
		ch, err := clickhouse.New(i.ctx, clickhouse.Config{
			Addr:     connURL,
			DB:       "test_db",
			Username: "su",
			Password: "su",
		})
		if err != nil {
			return err
		}
		defer ch.Close()

		clickhouseAddr = connURL
		return ch.Migrate(i.ctx)
	})
	i.Require().NoError(err)

	createTopicResponses, err := i.prepareTopics(ctx, defaultTopics...)
	i.Assert().NoError(err)

	for _, response := range createTopicResponses {
		i.Require().NoError(response.Err)
	}
	i.kafkaCLi.Close()

	i.app = app.New(func() (config.Config, error) {
		return config.Config{
			LogLevel: string(app.DEBUG),
			Clickhouse: clickhouse.Config{
				Addr:     clickhouseAddr,
				DB:       "test_db",
				Username: "su",
				Password: "su",
			},
			EventWorker: worker.Config{
				NumWorkers: 500,
			},
			EventConsumer: consumer.Config{
				Brokers:       []string{i.broker},
				ConsumerGroup: "event-cg",
				Topics:        []string{eventTopic},
				RetryCount:    5,
			},
			EventProducer: producer.Config{
				RetryAttempts: 5,
				RetryDelay:    time.Second * 10,
				SleepDuration: time.Second * 20,
				Brokers:       []string{i.broker},
				Topic:         eventTopic,
			},
		}, nil
	})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	i.wg = wg
	go func() {
		defer wg.Done()
		i.app.Start()
	}()

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Timeout = 5 * time.Second
	if _, err := retryClient.Get(fmt.Sprintf("%s/_/ready", fmt.Sprintf("http://localhost%s", defaultAddrPublic))); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func (i *IntegrationTestSuite) TearDownSuite() {
	i.cancelFn()
	i.wg.Wait()
	err := i.redpandaContainer.Purge()
	err = i.clickhouseContainer.Purge()
	i.Assert().NoError(err)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (i *IntegrationTestSuite) prepareTopics(ctx context.Context, topics ...string) (kadm.CreateTopicResponses, error) {
	resp, err := i.kafkaCLi.CreateTopics(
		ctx,
		1,
		1,
		map[string]*string{},
		topics...,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (i *IntegrationTestSuite) cleanTopics(ctx context.Context, topics ...string) (kadm.DeleteTopicResponses, error) {
	return i.kafkaCLi.DeleteTopics(ctx, topics...)
}
