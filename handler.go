package pobj

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/picosh/pobj/storage"
	"github.com/picosh/send/send/utils"
)

type ctxBucketKey struct{}

func getBucket(ctx ssh.Context) (storage.Bucket, error) {
	bucket := ctx.Value(ctxBucketKey{}).(storage.Bucket)
	if bucket.Name == "" {
		return bucket, fmt.Errorf("bucket not set on `ssh.Context()` for connection")
	}
	return bucket, nil
}
func setBucket(ctx ssh.Context, bucket storage.Bucket) {
	ctx.SetValue(ctxBucketKey{}, bucket)
}

type FileData struct {
	*utils.FileEntry
	Text   []byte
	User   string
	Bucket storage.Bucket
}

type Config struct {
	Logger     *slog.Logger
	Storage    storage.ObjectStorage
	AssetNames AssetNames
}

type UploadAssetHandler struct {
	Cfg *Config
}

var _ utils.CopyFromClientHandler = &UploadAssetHandler{}
var _ utils.CopyFromClientHandler = (*UploadAssetHandler)(nil)

func NewUploadAssetHandler(cfg *Config) *UploadAssetHandler {
	if cfg.AssetNames == nil {
		cfg.AssetNames = &AssetNamesBasic{}
	}

	return &UploadAssetHandler{
		Cfg: cfg,
	}
}

func (h *UploadAssetHandler) GetLogger() *slog.Logger {
	return h.Cfg.Logger
}

func (h *UploadAssetHandler) Read(s ssh.Session, entry *utils.FileEntry) (os.FileInfo, utils.ReaderAtCloser, error) {
	fileInfo := &utils.VirtualFile{
		FName:    filepath.Base(entry.Filepath),
		FIsDir:   false,
		FSize:    entry.Size,
		FModTime: time.Unix(entry.Mtime, 0),
	}

	bucketName := h.Cfg.AssetNames.BucketName(s.User())
	bucket, err := h.Cfg.Storage.GetBucket(bucketName)
	if err != nil {
		return nil, nil, err
	}

	fname := h.Cfg.AssetNames.ObjectName(entry)
	contents, size, modTime, err := h.Cfg.Storage.GetObject(bucket, fname)
	if err != nil {
		return nil, nil, err
	}

	fileInfo.FSize = size
	fileInfo.FModTime = modTime

	reader := NewAllReaderAt(contents)

	return fileInfo, reader, nil
}

func (h *UploadAssetHandler) List(s ssh.Session, fpath string, isDir bool, recursive bool) ([]os.FileInfo, error) {
	var fileList []os.FileInfo
	userName := s.User()

	cleanFilename := fpath

	bucketName := h.Cfg.AssetNames.BucketName(userName)
	bucket, err := h.Cfg.Storage.GetBucket(bucketName)
	if err != nil {
		return fileList, err
	}

	if cleanFilename == "" || cleanFilename == "." {
		name := cleanFilename
		if name == "" {
			name = "/"
		}

		info := &utils.VirtualFile{
			FName:  name,
			FIsDir: true,
		}

		fileList = append(fileList, info)
	} else {
		if cleanFilename != "/" && isDir {
			cleanFilename += "/"
		}

		foundList, err := h.Cfg.Storage.ListObjects(bucket, cleanFilename, recursive)
		if err != nil {
			return fileList, err
		}

		fileList = append(fileList, foundList...)
	}

	return fileList, nil
}

func (h *UploadAssetHandler) Validate(s ssh.Session) error {
	var err error
	userName := s.User()

	assetBucket := h.Cfg.AssetNames.BucketName(userName)
	bucket, err := h.Cfg.Storage.UpsertBucket(assetBucket)
	if err != nil {
		return err
	}
	setBucket(s.Context(), bucket)

	pk, _ := utils.KeyText(s)
	h.Cfg.Logger.Info(
		"attempting to upload files",
		"user", userName,
		"bucket", bucket.Name,
		"publicKey", pk,
	)
	return nil
}

func (h *UploadAssetHandler) Write(s ssh.Session, entry *utils.FileEntry) (string, error) {
	var origText []byte
	if b, err := io.ReadAll(entry.Reader); err == nil {
		origText = b
	}
	fileSize := binary.Size(origText)
	// TODO: hack for now until I figure out how to get correct
	// filesize from sftp,scp,rsync
	entry.Size = int64(fileSize)
	userName := s.User()

	bucket, err := getBucket(s.Context())
	if err != nil {
		h.Cfg.Logger.Error(err.Error())
		return "", err
	}

	data := &FileData{
		FileEntry: entry,
		User:      userName,
		Text:      origText,
		Bucket:    bucket,
	}
	err = h.writeAsset(data)
	if err != nil {
		h.Cfg.Logger.Error(err.Error())
		return "", err
	}

	// TODO: make it the object store URL
	url := fmt.Sprintf("%s%s", bucket.Name, h.Cfg.AssetNames.ObjectName(entry))
	return url, nil
}

func (h *UploadAssetHandler) validateAsset(data *FileData) (bool, error) {
	return true, nil
}

func (h *UploadAssetHandler) writeAsset(data *FileData) error {
	valid, err := h.validateAsset(data)
	if !valid {
		return err
	}

	objectFileName := h.Cfg.AssetNames.ObjectName(data.FileEntry)
	if data.Size == 0 {
		err = h.Cfg.Storage.DeleteObject(data.Bucket, objectFileName)
		if err != nil {
			return err
		}
	} else {
		reader := bytes.NewReader(data.Text)

		h.Cfg.Logger.Info(
			"uploading file to bucket",
			"user",
			data.User,
			"bucket",
			data.Bucket.Name,
			"object",
			objectFileName,
		)

		_, err = h.Cfg.Storage.PutObject(
			data.Bucket,
			objectFileName,
			utils.NopReaderAtCloser(reader),
			data.FileEntry,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
