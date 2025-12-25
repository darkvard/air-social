package email

import (
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
)

func consumeLoop(
	ctx context.Context,
	msgs <-chan amqp.Delivery,
	disp domain.EventHandler,
	done <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			handleMessage(ctx, msg, disp)
		}
	}
}

func handleMessage(
	ctx context.Context,
	msg amqp.Delivery,
	disp domain.EventHandler,
) {
	var evt domain.EventPayload
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		msg.Nack(false, false)
		return
	}

	if err := disp.Handle(ctx, evt); err != nil {
		handleRetry(msg)
		return
	}

	msg.Ack(false)
}
