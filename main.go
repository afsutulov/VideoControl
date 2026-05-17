package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to JSON configuration file")
	flag.StringVar(configPath, "c", defaultConfigPath, "path to JSON configuration file")
	run := flag.Bool("run", false, "run video monitoring service")
	version := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *version {
		fmt.Println(AppName, AppVersion)
		return
	}

	if !*run {
		printWelcome()
		return
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	if err := initLogging(cfg.Logging); err != nil {
		log.Fatalf("logging error: %v", err)
	}

	logInfo("%s %s started", AppName, AppVersion)
	logInfo("config loaded: %s", *configPath)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := NewApp(cfg)
	if err := app.Run(ctx); err != nil {
		logError("service stopped with error: %v", err)
		os.Exit(1)
	}
	logInfo("%s stopped", AppName)
}

// printWelcome показывает краткое описание и доступные ключи.
func printWelcome() {
	fmt.Printf("%s %s\n", AppName, AppVersion)
	fmt.Println("Lightweight RTSP motion recorder with optional WebDAV upload.")
	fmt.Println("Designed as a small companion service for SmartHome/MQTT installations.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  VideoControl --run --config /etc/videocontrol.conf")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --run              run video monitoring service")
	fmt.Println("  --config, -c PATH  path to config file, default /etc/videocontrol.conf")
	fmt.Println("  --version          show version")
}
