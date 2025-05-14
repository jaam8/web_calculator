package main

import (
	"context"
	"github.com/jaam8/web_calculator/auth_service/internal/config"
	"github.com/jaam8/web_calculator/auth_service/internal/ports/adapters/cache"
	"github.com/jaam8/web_calculator/auth_service/internal/ports/adapters/storage"
	"github.com/jaam8/web_calculator/auth_service/internal/server"
	"github.com/jaam8/web_calculator/common-lib/logger"
	"github.com/jaam8/web_calculator/common-lib/postgres"
	"github.com/jaam8/web_calculator/common-lib/redis"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx = context.WithValue(ctx, "log_level", cfg.LogLevel)
	ctx, _ = logger.New(ctx)

	authCfg := cfg.AuthService
	redisCfg := cfg.Redis
	postgresCfg := cfg.Postgres

	redisClient, err := redis.NewRedisClient(ctx, redisCfg, authCfg.RedisDB)
	if err != nil {
		log.Fatalf("failed to create redis client: %v", err)
	}

	PostgresClient, err := postgres.New(ctx, postgresCfg)
	defer PostgresClient.Close()
	if err != nil {
		log.Fatalf("failed to create postgres client: %v", err)
	}

	err = postgres.Migrate(ctx, postgresCfg, cfg.MigrationPath)
	if err != nil {
		log.Fatalf("failed to migrate postgres: %v", err)
	}
	redisAdapter := cache.NewAuthCacheAdapter(redisClient,
		time.Hour*time.Duration(cfg.RefreshExpiration),
		time.Minute*time.Duration(cfg.AccessExpiration),
	)

	postgresAdapter := storage.NewAuthPostgresAdapter(PostgresClient)

	Server := server.NewAuthService(postgresAdapter, redisAdapter, cfg.JwtSecret,
		time.Hour*time.Duration(cfg.RefreshExpiration),
		time.Minute*time.Duration(cfg.AccessExpiration),
	)

	grpcServer, err := server.CreateGRPC(Server)
	if err != nil {
		log.Fatalf("failed to create gRPC server: %v", err)
	}

	go server.RunGRPC(ctx, grpcServer, authCfg.Port)

	<-ctx.Done()
	grpcServer.GracefulStop()
	stop()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "AUTH_SERVICE server stopped")
}
