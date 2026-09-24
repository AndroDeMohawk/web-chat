package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageHandler func(payload MessagePayload)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic string, groupId string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupId,
		MinBytes: 10,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: reader}
}

func (c *Consumer) Start(ctx context.Context, handler MessageHandler, logger *zap.Logger) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("kafka consumer stopped")
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					logger.Error("kafka reader error", zap.Error(err))
					continue
				}
				var payload MessagePayload
				if err := json.Unmarshal(msg.Value, &payload); err != nil {
					logger.Error("json unmarshal error", zap.Error(err))
					continue
				}
				handler(payload)
			}

		}
	}()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
