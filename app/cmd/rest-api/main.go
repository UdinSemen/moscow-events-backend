package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/config"
	server "github.com/UdinSemen/moscow-events-backend/internal/http-server"
	"github.com/UdinSemen/moscow-events-backend/internal/http-server/handlers"
	jwt_manager "github.com/UdinSemen/moscow-events-backend/internal/jwt-manager"
	"github.com/UdinSemen/moscow-events-backend/internal/services"
	storage "github.com/UdinSemen/moscow-events-backend/internal/storage/postgres"
	redis "github.com/UdinSemen/moscow-events-backend/internal/storage/redis"
	"github.com/UdinSemen/moscow-events-backend/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	dev     = "dev"
	local   = "local"
	prod    = "prod"
	timeOut = 5
)

func main() {
	cfg := config.MustLoad()
	var zapLevel zapcore.Level
	switch cfg.Env {
	case dev:
		zapLevel = zapcore.DebugLevel
	case local:
		zapLevel = zapcore.DebugLevel
	case prod:
		zapLevel = zapcore.WarnLevel
	}
	logger, err := utils.CreateLogger(zapLevel)
	defer func() {
		if err = logger.Sync(); err != nil {
			logger.Sugar().Error(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	tokenManager, err := jwt_manager.NewManager(cfg.Jwt.SecretKey, &cfg.Jwt.AccessTokenTTL)
	if err != nil {
		zap.S().Fatalf(err.Error())
	}

	redisStorage := redis.NewRedisClient(cfg)
	if err := redisStorage.Ping(context.Background()); err != nil {
		zap.S().Fatalf(err.Error())
	}
	postgresStorage, err := storage.InitPgStorage(cfg)
	if err := postgresStorage.Ping(); err != nil {
		zap.S().Fatalf(err.Error())
	}

	if err != nil {
		zap.S().Fatalf(err.Error())
	}
	service := services.NewService(redisStorage,
		postgresStorage,
		cfg.Jwt.RefreshTokenTTL,
		tokenManager)
	handler := handlers.NewHandler(service, tokenManager)

	srv := new(server.Server)
	go func() {
		if err := srv.Run(cfg, handler.InitRoutes()); err != nil && !errors.Is(http.ErrServerClosed, err) {
			zap.S().Panicf("listen %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), timeOut*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.S().Infof("Server Shutdown: %s", err)
	}
	if err := redisStorage.Close(); err != nil {
		zap.S().Errorf("Error with closing redis %s", err)
	}

	select {
	case <-ctx.Done():
		zap.S().Infof("Timeout of %d seconds", timeOut)
	}
	zap.S().Info("Server exiting")
}
