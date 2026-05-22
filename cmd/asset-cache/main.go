package main

import (
	"log"

	"github.com/auto-forge-org/asset-cache/config"
	"github.com/auto-forge-org/asset-cache/internal/api"
	"github.com/auto-forge-org/asset-cache/internal/cache"
	"github.com/auto-forge-org/asset-cache/internal/service"
	"github.com/auto-forge-org/asset-cache/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	store := storage.NewMemoryStore()
	lru := cache.NewLRU(cfg.CacheSize)
	svc := service.NewAssetService(store, lru, cfg.SignerKey)

	r := gin.Default()
	api.NewHandler(svc).Register(r)

	addr := ":" + cfg.Port
	log.Printf("asset-cache listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
