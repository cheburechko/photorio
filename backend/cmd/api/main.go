package main

import (
	"backend/internal"
	"flag"
	"fmt"
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

type ()

func readConfig() (*internal.AppConfig, error) {
	var cfg internal.AppConfig

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

	app, err := internal.NewApp(cfg)

	if err != nil {
		slog.Error("Failed to create app", slog.String("error", err.Error()))
		return
	}
	defer app.Close()

	router := internal.BuildRouter(app)

	addr := fmt.Sprintf(":%s", app.Config.Port)

	err = router.Run(addr)
	slog.Error("Finishing", slog.String("error", err.Error()))
}
