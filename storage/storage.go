package storage

import (
	"os"
	"time"

	"github.com/picosh/send/send/utils"
)

type Bucket struct {
	Name string
	Path string
	Root string
}

type ObjectStorage interface {
	GetBucket(name string) (Bucket, error)
	UpsertBucket(name string) (Bucket, error)
	ListBuckets() ([]string, error)

	DeleteBucket(bucket Bucket) error
	GetBucketQuota(bucket Bucket) (uint64, error)
	GetObjectSize(bucket Bucket, fpath string) (int64, error)
	GetObject(bucket Bucket, fpath string) (utils.ReaderAtCloser, int64, time.Time, error)
	PutObject(bucket Bucket, fpath string, contents utils.ReaderAtCloser, entry *utils.FileEntry) (string, error)
	DeleteObject(bucket Bucket, fpath string) error
	ListObjects(bucket Bucket, dir string, recursive bool) ([]os.FileInfo, error)
}
