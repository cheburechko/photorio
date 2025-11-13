package main

import (
	"backend/internal"
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Config struct {
		Postgres   internal.PostgresConfig
		Kafka      internal.KafkaConfig
		PollPeriod time.Duration `yaml:"poll_period"`
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

	producer, err := internal.NewProducer(cfg.Kafka)
	if err != nil {
		slog.Error("Failed to create producer", slog.String("error", err.Error()))
		return
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.Group, saramaConfig)
	if err != nil {
		slog.Error("Failed to create consumer group", slog.String("error", err.Error()))
		return
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.Postgres.ConnectionUrl)
	if err != nil {
		slog.Error("Failed to create postgres", slog.String("error", err.Error()))
		return
	}

	consumer := internal.NewConsumer(cfg.Kafka, producer, dbpool)
	poller := internal.NewPoller(producer, dbpool, cfg.PollPeriod)

	ctx := context.Background()

	go poller.Run(ctx)

	slog.Info("Starting consumer")
	err = consumer.Run(ctx, consumerGroup)
	if err != nil {
		slog.Error("Failed to run consumer", slog.String("error", err.Error()))
		return
	}
}
