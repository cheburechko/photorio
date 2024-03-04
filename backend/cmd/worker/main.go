package main

import (
	"backend/internal"
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Config struct {
		Postgres internal.PostgresConfig
		Kafka    internal.KafkaConfig
	}
)

func readConfig() (*Config, error) {
	var cfg Config

	configPath := flag.String("config", "config.yaml", "Path to config")

	flag.Parse()

	err := cleanenv.ReadConfig(*configPath, &cfg)

	return &cfg, err
}

func main() {
	cfg, err := readConfig()
	if err != nil {
		slog.Error("Failed to read config", slog.String("error", err.Error()))
		return
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, saramaConfig)
	if err != nil {
		slog.Error("Failed to create producer", slog.String("error", err.Error()))
		return
	}

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.Group, saramaConfig)
	if err != nil {
		slog.Error("Failed to create consumer group", slog.String("error", err.Error()))
		return
	}

	psqlUrl := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database)
	dbpool, err := pgxpool.New(context.Background(), psqlUrl)
	if err != nil {
		slog.Error("Failed to create postgres", slog.String("error", err.Error()))
		return
	}

	consumer := internal.NewConsumer(cfg.Kafka, producer, dbpool)

	ctx := context.Background()

	slog.Info("Starting consumer")
	err = consumer.Run(ctx, consumerGroup)
	if err != nil {
		slog.Error("Failed to run consumer", slog.String("error", err.Error()))
		return
	}
}
