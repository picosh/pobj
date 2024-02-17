package objx

import (
	"log/slog"
	"os"

	"github.com/picosh/objx/storage"
)

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func EnvDriverDetector(logger *slog.Logger) (storage.ObjectStorage, error) {
	driver := GetEnv("OBJECT_DRIVER", "fs")
	logger.Info("driver detected", "driver", driver)

	if driver == "minio" {
		url := GetEnv("OBJECT_URL", "")
		user := GetEnv("OBJECT_USER", "")
		pass := GetEnv("OBJECT_PASS", "")
		logger.Info(
			"object config detected",
			"url", url,
			"user", user,
		)
		return storage.NewStorageMinio(
			url,
			user,
			pass,
		)
	}

	storageDir := GetEnv("OBJECT_URL", "./.storage")
	logger.Info("object config detected", "dir", storageDir)
	return storage.NewStorageFS(storageDir)
}
