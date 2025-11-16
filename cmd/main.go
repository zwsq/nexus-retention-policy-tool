package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nexus-retention-policy/internal/config"
	"nexus-retention-policy/internal/logger"
	"nexus-retention-policy/internal/nexus"
	"nexus-retention-policy/internal/retention"

	"github.com/robfig/cron/v3"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("🚀 Nexus Retention Policy Tool")
	fmt.Println("================================")

	// Initialize logger
	log, err := logger.NewLogger(cfg.LogFile)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Close()

	// Initialize Nexus client
	client := nexus.NewClient(cfg.Nexus.URL, cfg.Nexus.Username, cfg.Nexus.Password, cfg.Nexus.Timeout)

	// Initialize policy engine
	engine := retention.NewPolicyEngine(client, cfg, log)

	// Check if scheduling is enabled
	if cfg.Schedule == "" {
		// One-time execution
		fmt.Println("Mode: One-time execution")
		return engine.Execute()
	}

	// Scheduled execution
	fmt.Printf("Mode: Scheduled execution (%s)\n", cfg.Schedule)
	fmt.Println("Press Ctrl+C to stop")

	c := cron.New()
	_, err = c.AddFunc(cfg.Schedule, func() {
		fmt.Printf("\n⏰ Scheduled execution started at %s\n", formatTime())
		if err := engine.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		}
		fmt.Printf("⏰ Scheduled execution completed at %s\n", formatTime())
	})

	if err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}

	c.Start()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n\n👋 Shutting down gracefully...")
	c.Stop()

	return nil
}

func formatTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
