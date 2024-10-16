package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/picosh/send/utils"
)

type StorageMemory struct {
	storage map[string]map[string]string
}

var _ ObjectStorage = &StorageMemory{}
var _ ObjectStorage = (*StorageMemory)(nil)

func NewStorageMemory() (*StorageMemory, error) {
	return &StorageMemory{
		storage: map[string]map[string]string{},
	}, nil
}

func (s *StorageMemory) GetBucket(name string) (Bucket, error) {
	bucket := Bucket{
		Name: name,
		Path: name,
	}

	_, ok := s.storage[name]
	if !ok {
		return bucket, fmt.Errorf("bucket does not exist")
	}

	return bucket, nil
}

func (s *StorageMemory) UpsertBucket(name string) (Bucket, error) {
	bucket, err := s.GetBucket(name)
	if err == nil {
		return bucket, nil
	}

	s.storage[name] = map[string]string{}
	return bucket, nil
}

func (s *StorageMemory) GetBucketQuota(bucket Bucket) (uint64, error) {
	objects := s.storage[bucket.Path]
	size := 0
	for _, val := range objects {
		size += len(val)
	}
	return uint64(size), nil
}

func (s *StorageMemory) DeleteBucket(bucket Bucket) error {
	delete(s.storage, bucket.Path)
	return nil
}

func (s *StorageMemory) GetObject(bucket Bucket, fpath string) (utils.ReaderAtCloser, *ObjectInfo, error) {
	objInfo := &ObjectInfo{
		LastModified: time.Time{},
		Metadata:     nil,
		UserMetadata: map[string]string{},
	}

	dat, ok := s.storage[bucket.Path][fpath]
	if !ok {
		return nil, objInfo, fmt.Errorf("object does not exist: %s", fpath)
	}

	objInfo.Size = int64(len(dat))
	reader := utils.NopReaderAtCloser(strings.NewReader(dat))
	return reader, objInfo, nil
}

func (s *StorageMemory) PutObject(bucket Bucket, fpath string, contents io.Reader, entry *utils.FileEntry) (string, int64, error) {
	buf := new(strings.Builder)
	size, err := io.Copy(buf, contents)
	if err != nil {
		return "", 0, err
	}
	s.storage[bucket.Path][fpath] = buf.String()
	return fmt.Sprintf("%s%s", bucket.Path, fpath), size, nil
}

func (s *StorageMemory) DeleteObject(bucket Bucket, fpath string) error {
	delete(s.storage[bucket.Path], fpath)
	return nil
}

func (s *StorageMemory) ListBuckets() ([]string, error) {
	buckets := []string{}
	for key := range s.storage {
		buckets = append(buckets, key)
	}
	return buckets, nil
}

func (s *StorageMemory) ListObjects(bucket Bucket, dir string, recursive bool) ([]os.FileInfo, error) {
	var fileList []os.FileInfo
	resolved := dir

	objects := s.storage[bucket.Path]
	// dir is actually an object
	oval, ok := objects[resolved]
	if ok {
		fileList = append(fileList, &utils.VirtualFile{
			FName:    filepath.Base(resolved),
			FIsDir:   false,
			FSize:    int64(len(oval)),
			FModTime: time.Time{},
		})
		return fileList, nil
	}

	for key, val := range objects {
		if !strings.HasPrefix(key, resolved) {
			continue
		}

		rep := strings.Replace(key, resolved, "", 1)
		fdir := filepath.Dir(rep)
		fname := filepath.Base(rep)
		paths := strings.Split(fdir, "/")

		if fdir == "/" {
			ffname := filepath.Base(resolved)
			fileList = append(fileList, &utils.VirtualFile{
				FName:  ffname,
				FIsDir: true,
			})
		}

		for _, p := range paths {
			if p == "" || p == "/" || p == "." {
				continue
			}
			fileList = append(fileList, &utils.VirtualFile{
				FName:  p,
				FIsDir: true,
			})
		}

		trimRes := strings.TrimSuffix(resolved, "/")
		dirKey := filepath.Dir(key)
		if recursive {
			fileList = append(fileList, &utils.VirtualFile{
				FName:    fname,
				FIsDir:   false,
				FSize:    int64(len(val)),
				FModTime: time.Time{},
			})
		} else if resolved == dirKey || trimRes == dirKey {
			fileList = append(fileList, &utils.VirtualFile{
				FName:    fname,
				FIsDir:   false,
				FSize:    int64(len(val)),
				FModTime: time.Time{},
			})
		}
	}

	return fileList, nil
}
