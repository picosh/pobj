package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/wish"
	"github.com/picosh/pobj"
)

func main() {
	logger := slog.Default()
	host := pobj.GetEnv("SSH_HOST", "0.0.0.0")
	port := pobj.GetEnv("SSH_PORT", "2222")
	keyPath := pobj.GetEnv("SSH_AUTHORIZED_KEYS", "./ssh_data/authorized_keys")
	bucketName := pobj.GetEnv("OBJECT_BUCKET_NAME", "")

	st, err := pobj.EnvDriverDetector(logger)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	cfg := &pobj.Config{
		Logger:  logger,
		Storage: st,
	}
	if bucketName != "" {
		cfg.AssetNames = &pobj.AssetNamesForceBucket{
			Name: bucketName,
		}
	}

	handler := pobj.NewUploadAssetHandler(cfg)

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithHostKeyPath("ssh_data/term_info_ed25519"),
		wish.WithAuthorizedKeys(keyPath),
		pobj.WithProxy(handler),
	)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Info(
		"starting SSH server",
		"host", host,
		"port", port,
	)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			logger.Error(err.Error())
		}
	}()

	<-done
	logger.Info("stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()
	if err := s.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
}
