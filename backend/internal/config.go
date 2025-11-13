package internal

import "github.com/elastic/go-elasticsearch/v8"

type (
	PostgresConfig struct {
		ConnectionUrl     string `yaml:"connection_url"`
	}

	KafkaConfig struct {
		Brokers      []string `yaml:"brokers"`
		Group        string   `yaml:"group"`
		TaskTopic    string   `yaml:"task_topic"`
		CaptionTopic string   `yaml:"caption_topic"`
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
