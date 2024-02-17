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
		url := GetEnv("MINIO_URL", "")
		user := GetEnv("MINIO_ROOT_USER", "")
		pass := GetEnv("MINIO_ROOT_PASSWORD", "")
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
