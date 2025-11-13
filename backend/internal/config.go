package internal

import (
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type (
	PostgresConfig struct {
		ConnectionUrl     string `env:"POSTGRES_CONNECTION_URL"`
	}

	KafkaConfig struct {
		Brokers      []string `env:"KAFKA_BROKERS"`
		Group        string   `yaml:"group"`
		TaskTopic    string   `yaml:"task_topic" env:"TASK_TOPIC"`
		CaptionTopic string   `yaml:"caption_topic" env:"CAPTION_TOPIC"`
	}

	AppConfig struct {
		Port          string               `yaml:"port" env:"PORT" env-default:"9000"`
		TemplateGLOB  string               `yaml:"template_glob"`
		Elasticsearch elasticsearch.Config `yaml:"elasticsearch"`
		CookieSecret  string               `env:"COOKIE_SECRET"`
		Kafka         KafkaConfig          `yaml:"kafka"`
		Postgres      PostgresConfig
	}
)

func (c *AppConfig) Update() error {
	if os.Getenv("ELASTICSEARCH_ADDRESSES") != "" {
		c.Elasticsearch.Addresses = strings.Split(os.Getenv("ELASTICSEARCH_ADDRESSES"), ",")
	}
	return nil
}
