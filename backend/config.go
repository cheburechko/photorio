package main

import (
	"flag"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	PostgresConfig struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
		Database string `yaml:"database"`
	}

	Config struct {
		Port          string               `yaml:"port" env:"PORT" env-default:"9000"`
		TemplateGLOB  string               `yaml:"template_glob"`
		Elasticsearch elasticsearch.Config `yaml:"elasticsearch"`
		Postgres      PostgresConfig
	}
)

func readConfig() (*Config, error) {
	var cfg Config

	configPath := flag.String("config", "config.yaml", "Path to config")

	flag.Parse()

	err := cleanenv.ReadConfig(*configPath, &cfg)

	return &cfg, err
}
