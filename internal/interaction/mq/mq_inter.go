package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	mess "air-social/internal/infrastructure/messaging"
	"air-social/internal/worker"
	"air-social/pkg"
)

const (
	interaction = "INTERACTION"
	publisher   = "PUBLISHER"
	consumer    = "CONSUMER"
	timeout     = "TIMEOUT"
)

type rabbitMQ struct {
	publisher *mess.Publisher
	workerMgr *worker.Manager
}

func newRabbitMQ(conn *amqp091.Connection) *rabbitMQ {
	mgr := worker.NewManager(
		worker.NewEmailWorker(
			conn,
			mess.EventsExchange,
			mess.QueueConfig{
				Queue:      "email.interaction.q",
				RoutingKey: "email.*",
			},
			newEventHandler(),
		),
	)

	pub, err := mess.NewPublisher(
		conn,
		mess.EventsExchange,
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

func TestRabbitMQ(conn *amqp091.Connection) {
	mq := newRabbitMQ(conn)
	mq.testing()
}

func (r *rabbitMQ) testing() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go r.startWorker(ctx)
	time.Sleep(time.Second)

	r.messageHandle(ctx)
	r.stopWorker()
}

func (r *rabbitMQ) startWorker(ctx context.Context) {
	logInfo(interaction, "STARTING WORKER", "...")
	if err := r.workerMgr.Start(ctx); err != nil {
		logError(interaction, "WORKER STOPPED", "Error: %v", err)
	}
}

func (r *rabbitMQ) messageHandle(ctx context.Context) {
	for _, c := range messCases {
		pkg.Log().Info("--------------------------------------------------------------------------------")
		func() {
			evt := initEvent(fmt.Sprintf("test-%d", time.Now().UnixNano()), c.key)

			pubCtx, pubCancel := context.WithTimeout(ctx, 5*time.Second)
			// Important: Defer is function-scoped. We use an anonymous function to ensure
			// cleanup runs at the end of each iteration, preventing resource leaks.
			defer pubCancel()

			if c.name == connErrState {
				logInfo(timeout, "Simulating", "Closing publisher connection...")
				r.publisher.Close()
				time.Sleep(100 * time.Millisecond)
			}

			if c.name == timeoutState {
				pubCtx, pubCancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer pubCancel()
			}

			if err := r.publisher.Publish(pubCtx, c.key, evt); err != nil {
				logError(publisher, "Publish failed", "Error: %v", err)
			} else {
				logInfo(publisher, "Publish success", "Target: %s", c.key)
			}
		}()

		time.Sleep(800 * time.Millisecond)
	}

}

func initEvent(name, key string) domain.EventPayload {
	return domain.EventPayload{
		EventID:   name,
		EventType: key,
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{}`),
	}
}

func (r *rabbitMQ) stopWorker() {
	r.workerMgr.Stop()
}
