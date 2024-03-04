package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Consumer struct {
	config   KafkaConfig
	producer *Producer
	postgres *pgxpool.Pool
}

func NewConsumer(config KafkaConfig, producer *Producer, postgres *pgxpool.Pool) *Consumer {
	return &Consumer{
		config:   config,
		producer: producer,
		postgres: postgres,
	}
}

func (consumer *Consumer) Run(ctx context.Context, consumerGroup sarama.ConsumerGroup) error {
	for {
		if err := consumerGroup.Consume(ctx, []string{consumer.config.TaskTopic}, consumer); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}
			slog.Error("Error from consumer", slog.String("error", err.Error()))
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				slog.Info("message channel was closed")
				return nil
			}
			var err error
			if message.Topic == consumer.config.TaskTopic {
				err = consumer.ProcessTask(session.Context(), message.Value)
			} else {
				slog.Error("Unknown topic", slog.String("topic", message.Topic))
			}
			if err != nil {
				slog.Error("Error processing message", slog.String("error", err.Error()))
				return err
			} else {
				session.MarkMessage(message, "")
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

func (consumer *Consumer) ProcessTask(ctx context.Context, value []byte) error {
	var event CaptionTaskEvent

	if err := json.Unmarshal(value, &event); err != nil {
		return err
	}

	slog.Info("Processing task", slog.Int("id", event.Id), slog.String("prompt", event.Prompt), slog.String("url", event.Url))

	tx, err := consumer.postgres.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var completed int
	var total int

	if event.Url == "" {
		slog.Info("Creating new task", slog.String("prompt", event.Prompt))
		_, err = tx.Exec(
			ctx,
			`update caption_tasks
				set total_subtasks = 100, 
					completed_subtasks = 0, 
					status = 'scheduling'
				where id = $1`,
			event.Id,
		)
		if err != nil {
			return err
		}
		total = 100
		completed = 0
	} else {
		slog.Info("Completing subtask", slog.String("prompt", event.Prompt))
		err = tx.QueryRow(
			ctx,
			`update caption_tasks 
				set completed_subtasks = completed_subtasks + 1,
					status = CASE completed_subtasks + 1
						WHEN total_subtasks THEN 'completed'::caption_status
						ELSE 'scheduling'::caption_status
					END
			where id = $1 AND completed_subtasks < total_subtasks
			returning total_subtasks, completed_subtasks`,
			event.Id,
		).Scan(&total, &completed)
		if err != nil {
			return err
		}
	}

	if completed < total {
		slog.Info("Sending task to kafka", slog.String("prompt", event.Prompt), slog.Int("completed", completed), slog.Int("total", total))

		event.Url = "ololo"
		err = consumer.producer.SendTaskEvent(&event)

		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	slog.Info("Task processed", slog.String("prompt", event.Prompt))

	return nil
}
