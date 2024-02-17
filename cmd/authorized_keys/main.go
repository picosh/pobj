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
	"github.com/picosh/objx/storage"
)

func main() {
	host := objx.GetEnv("SSH_HOST", "0.0.0.0")
	port := objx.GetEnv("SSH_PORT", "2222")
	keyPath := objx.GetEnv("SSH_AUTHORIZED_KEYS", "./ssh_data/authorized_keys")
	storageDir := "./.storage"
	logger := slog.Default()
	storage, err := storage.NewStorageFS(storageDir)

	if err != nil {
		logger.Error(err.Error())
		return
	}

	cfg := &objx.Config{
		Logger:  logger,
		Storage: storage,
	}

	handler := objx.NewUploadAssetHandler(cfg)

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithHostKeyPath("ssh_data/term_info_ed25519"),
		wish.WithAuthorizedKeys(keyPath),
		objx.WithStorageProxy(handler),
	)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Info(
		"Starting SSH server",
		"host", host,
		"port", port,
	)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			logger.Error(err.Error())
		}
	}()

	<-done
	logger.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
}
