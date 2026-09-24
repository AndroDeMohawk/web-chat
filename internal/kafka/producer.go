package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type MessagePayload struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{
		writer: &writer,
	}
}
func (p *Producer) SendMessage(ctx context.Context, payload MessagePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %w", err)
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
