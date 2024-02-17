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
	"github.com/picosh/objx"
)

func main() {
	logger := slog.Default()
	host := objx.GetEnv("SSH_HOST", "0.0.0.0")
	port := objx.GetEnv("SSH_PORT", "2222")
	keyPath := objx.GetEnv("SSH_AUTHORIZED_KEYS", "./ssh_data/authorized_keys")

	st, err := objx.EnvDriverDetector(logger)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	cfg := &objx.Config{
		Logger:  logger,
		Storage: st,
	}

	handler := objx.NewUploadAssetHandler(cfg)

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithHostKeyPath("ssh_data/term_info_ed25519"),
		wish.WithAuthorizedKeys(keyPath),
		objx.WithProxy(handler),
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
