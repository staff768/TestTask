package main

import (
	"net/http"

	_ "testtask/docs"
	"testtask/internal/cache"
	"testtask/internal/config"
	"testtask/internal/repository"
	logger "testtask/pkg"
)

// @title TestTask Subscriptions API
// @version 1.0
// @description API for managing subscriptions with Redis caching
// @BasePath /

func main() {
	logger.Init()
	cfg, err := config.LoadFromYAML()
	logger.Log.Info("Loaded config")
	if err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.Connect(cfg.Database)
	logger.Log.Info("Connected to database")
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := cache.Connect(cfg.Redis)
	logger.Log.Info("Connected to redis")
	if err != nil {
		logger.Log.WithError(err).Warn("Redis not available; continuing without cache")
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	appRepo = repository.NewSubscriptionRepository(db, logger.Log, redisClient)

	logger.Log.Infof("Starting web server at %s", cfg.Server.Port)
	if err := http.ListenAndServe(cfg.Server.Port, routes()); err != nil {
		logger.Log.Error(err.Error())
	}
}
