package consumer

import (
	"context"
	"fmt"
	"sync"

	"air-social/internal/infrastructure/rabbitmq/config"
)

type Group struct {
	consumers []*Consumer
}

func NewGroup(deps Deps, queues []config.QueueConfig, domain string) *Group {
	items := make([]*Consumer, len(queues))

	for i, qCfg := range queues {
		d := deps
		d.QueueCfg = qCfg
		d.Domain = Domain(domain)
		items[i] = NewConsumer(d)
	}

	return &Group{consumers: items}
}

func (g *Group) Start(ctx context.Context, wg *sync.WaitGroup) error {
	for _, c := range g.consumers {
		if err := c.Start(ctx, wg); err != nil {
			return err
		}
	}
	return nil
}

func (g *Group) Stop() error {
	errMap := map[string]string{}

	for _, c := range g.consumers {
		if err := c.Stop(); err != nil {
			key := fmt.Sprintf("queue:%s-routingKey:%s", c.QueueCfg.Queue, c.QueueCfg.RoutingKey)
			errMap[key] = err.Error()
		}
	}

	if len(errMap) > 0 {
		return fmt.Errorf("stop consumer group errors: %+v", errMap)
	}

	return nil
}
