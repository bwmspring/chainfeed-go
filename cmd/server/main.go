package main

import (
	"flag"
	"log"

	_ "github.com/bwmspring/chainfeed-go/docs/swagger"
	"github.com/bwmspring/chainfeed-go/internal/app"
)

// @title           ChainFeed API
// @version         1.0
// @description     ChainFeed - 链上数据实时信息流平台
// @description     像刷 Twitter 一样追踪链上活动

// @contact.name   API Support
// @contact.email  bwm029@gmail.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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
