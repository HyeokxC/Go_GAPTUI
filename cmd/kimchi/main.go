package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/hyeokx/Go_GAPTUI/internal/background"
	"github.com/hyeokx/Go_GAPTUI/internal/config"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load("config.toml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	log.Info().
		Int("symbols", len(cfg.Symbols)).
		Bool("auto_discover", cfg.AutoDiscoverSymbols).
		Msg("config loaded")

	runner, err := background.NewRunner(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create runner")
	}

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Info().Msg("shutting down")
		cancel()
	}()

	if err := runner.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("runner failed")
	}
}
