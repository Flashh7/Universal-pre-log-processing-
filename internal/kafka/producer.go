package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sih/log-platform/internal/event"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           10 * time.Millisecond,
		AllowAutoTopicCreation: true,
	}

	return &Producer{
		writer: w,
	}
}

func (p *Producer) Publish(ctx context.Context, ev event.Event) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Value: data,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
