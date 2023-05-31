package worker

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/leshachaplin/datalog/internal/worker/redpanda/consumer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/producer"
	"github.com/leshachaplin/datalog/internal/worker/redpanda/testingh"
)

const (
	topic = "topic"
)

var (
	defaultTopics = []string{topic}
)

type IntegrationTestSuite struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	kafkaCLi  *kadm.Client
	container *testingh.Container
	broker    string

	consumerCfg consumer.Config
	producerCfg producer.Config
	*rand.Rand

	suite.Suite
}

func (i *IntegrationTestSuite) SetupSuite() {
	var err error
	ctx, cnsl := context.WithTimeout(context.Background(), time.Minute*2)
	i.ctx = ctx
	i.cancelFn = cnsl

	i.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	i.container, err = testingh.NewContainer(func(connURL string) error {
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

	createTopicResponses, err := i.prepareTopics(ctx, defaultTopics...)
	i.Assert().NoError(err)
	i.kafkaCLi.Close()

	for _, response := range createTopicResponses {
		i.Require().NoError(response.Err)
	}

	i.consumerCfg = consumer.Config{
		Brokers:       []string{i.broker},
		ConsumerGroup: "topic-cg",
		Topics:        []string{topic},
		RetryCount:    5,
	}
	i.producerCfg = producer.Config{
		RetryAttempts: 5,
		RetryDelay:    time.Second * 10,
		SleepDuration: time.Second * 20,
		Brokers:       []string{i.broker},
		Topic:         topic,
	}
}

func (i *IntegrationTestSuite) TearDownSuite() {
	i.cancelFn()
	err := i.container.Purge()
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
