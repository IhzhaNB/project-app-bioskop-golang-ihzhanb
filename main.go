// main.go
package main

import (
	"log"

	"cinema-booking/cmd"
	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/wire"
	"cinema-booking/pkg/database"
	"cinema-booking/pkg/utils"

	"go.uber.org/zap"
)

func main() {
	// Load config
	config, err := utils.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := utils.InitLogger(config.App.LogPath, config.App.Debug)
	if err != nil {
		log.Printf("Failed to init logger: %v. Using standard log.", err)
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	logger.Info("Starting application",
		zap.String("app", config.App.Name),
		zap.String("port", config.App.Port),
		zap.Bool("debug", config.App.Debug),
	)

	// Connect to database
	db, err := database.InitDB(config.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Initialize all repositories
	repos := repository.NewRepository(db, logger)

	// Wire all dependencies
	app := wire.Wiring(repos, config, logger)

	// Start server
	logger.Info("Starting HTTP server", zap.String("port", config.App.Port))

	cmd.APIServer(app.Router, config.App.Port)
}
