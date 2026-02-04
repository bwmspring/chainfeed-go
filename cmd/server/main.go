package main

import (
	"flag"
	"log"

	"chainfeed-go/internal/app"
)

var configPath = flag.String("config", "config/config.yaml", "path to config file")

func main() {
	flag.Parse()

	application, err := app.New(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
