package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/publisher"
	"air-social/internal/transport/worker"
	"air-social/pkg"
)

const (
	logInteraction = "INTERACTION"
	logPublisher   = "PUBLISHER"
	logConsumer    = "CONSUMER"
	logTimeout     = "TIMEOUT"
)

type rabbitMQ struct {
	publisher *publisher.Publisher
	workerMgr *worker.Manager
}

func newRabbitMQ(conn *amqp091.Connection, cache domain.CacheStorage) *rabbitMQ {
	mgr := worker.NewManager(
		worker.NewEmailWorker(
			conn,
			cache,
			newEventHandler(),
			config.EventsExchange,
			config.QueueConfig{
				Queue:      "email.interaction.q",
				RoutingKey: "email.*",
			},
		),
	)

	pub, err := publisher.NewPublisher(
		conn,
		config.EventsExchange,
		10,
	)
	if err != nil {
		panic(err)
	}

	return &rabbitMQ{
		publisher: pub,
		workerMgr: mgr,
	}
}

func TestRabbitMQ(conn *amqp091.Connection, c domain.CacheStorage) {
	mq := newRabbitMQ(conn, c)
	mq.testing()
}

func (r *rabbitMQ) testing() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go r.startWorker(ctx)
	time.Sleep(time.Second)

	r.messageHandle(ctx)
	r.stopWorker(ctx)
}

func (r *rabbitMQ) startWorker(ctx context.Context) {
	logInfo(logInteraction, "STARTING WORKER", "...")
	if err := r.workerMgr.Start(ctx); err != nil {
		logError(logInteraction, "WORKER STOPPED", "Error: %v", err)
	}
}

func (r *rabbitMQ) messageHandle(ctx context.Context) {
	for _, c := range messCases {
		pkg.Log().Info("--------------------------------------------------------------------------------")
		func() {
			evt := initEvent(fmt.Sprintf("test-%d", pkg.TimeNowUTC().UnixNano()), c.key)

			pubCtx, pubCancel := context.WithTimeout(ctx, 5*time.Second)
			// Important: Defer is function-scoped. We use an anonymous function to ensure
			// cleanup runs at the end of each iteration, preventing resource leaks.
			defer pubCancel()

			if c.name == connErrState {
				logInfo(logTimeout, "Simulating", "Closing publisher connection...")
				r.publisher.Close()
				time.Sleep(100 * time.Millisecond)
			}

			if c.name == timeoutState {
				pubCtx, pubCancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer pubCancel()
			}

			if err := r.publisher.Publish(pubCtx, c.key, evt); err != nil {
				logError(logPublisher, "Publish failed", "Error: %v", err)
			} else {
				logInfo(logPublisher, "Publish success", "Target: %s", c.key)
			}
		}()

		time.Sleep(800 * time.Millisecond)
	}

}

func initEvent(name, key string) domain.Event {
	return domain.Event{
		EventID:   name,
		EventType: domain.EventType(key),
		Timestamp: pkg.TimeNowUTC(),
		Data:      map[string]interface{}{},
	}
}

func (r *rabbitMQ) stopWorker(ctx context.Context) {
	r.workerMgr.Stop(ctx)
}
