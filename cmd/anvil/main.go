package main

import (
	"fmt"
	"os"

	"github.com/siddharth-bhatnagar/anvil/internal/config"
	"github.com/siddharth-bhatnagar/anvil/internal/tui"
	"github.com/siddharth-bhatnagar/anvil/internal/util"
)

func main() {
	// Initialize configuration
	configMgr := config.NewManager()
	if err := configMgr.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	cfg := configMgr.GetConfig()
	logConfig := util.LogConfig{
		LogDir:   cfg.LogDir,
		LogLevel: cfg.LogLevel,
	}

	if err := util.InitLogger(logConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Clean up old logs (keep last 7 days)
	if err := util.CleanupOldLogs(cfg.LogDir, 7); err != nil {
		util.Logger.Warn().Err(err).Msg("Failed to cleanup old logs")
	}

	util.Logger.Info().Msg("Starting Anvil")

	// Run TUI
	if err := tui.Run(); err != nil {
		util.Logger.Error().Err(err).Msg("TUI error")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	util.Logger.Info().Msg("Anvil stopped")
}
