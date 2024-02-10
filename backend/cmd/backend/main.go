package main

import (
	"fmt"
	"log/slog"
	"backend/internal"
)

func main() {
	cfg, err := internal.ReadConfig()
	if err != nil {
		slog.Error("Failed to read config", slog.String("error", err.Error()))
		return
	}

	app, err := internal.NewApp(cfg)

	if err != nil {
		slog.Error("Failed to create app", slog.String("error", err.Error()))
		return
	}

	router := internal.BuildRouter(app)

	addr := fmt.Sprintf(":%s", app.Config.Port)

	err = router.Run(addr)
	slog.Error("Finishing", slog.String("error", err.Error()))
}
