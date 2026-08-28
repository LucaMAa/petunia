package main

import (
	"log"
	"os"
	"petunia/internal/config"
	"petunia/internal/router"
)

func main() {
	cfg := config.LoadConfig()
	config.InitDB(cfg)
	config.InitRedis()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.Setup()

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
