package pobj

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/picosh/pobj/storage"
	"github.com/picosh/send/send/utils"
)

type ctxBucketKey struct{}

func getBucket(ctx ssh.Context) (storage.Bucket, error) {
	bucket, ok := ctx.Value(ctxBucketKey{}).(storage.Bucket)
	if !ok {
		return bucket, fmt.Errorf("bucket not set on `ssh.Context()` for connection")
	}
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

func (h *UploadAssetHandler) Delete(s ssh.Session, entry *utils.FileEntry) error {
	h.Cfg.Logger.Info("deleting file", "file", entry.Filepath)
	bucket, err := getBucket(s.Context())
	if err != nil {
		h.Cfg.Logger.Error(err.Error())
		return err
	}

	objectFileName, err := h.Cfg.AssetNames.ObjectName(s, entry)
	if err != nil {
		return err
	}
	return h.Cfg.Storage.DeleteObject(bucket, objectFileName)
}

func (h *UploadAssetHandler) Read(s ssh.Session, entry *utils.FileEntry) (os.FileInfo, utils.ReaderAtCloser, error) {
	fileInfo := &utils.VirtualFile{
		FName:    filepath.Base(entry.Filepath),
		FIsDir:   false,
		FSize:    entry.Size,
		FModTime: time.Unix(entry.Mtime, 0),
	}
	h.Cfg.Logger.Info("reading file", "file", fileInfo)

	bucketName, err := h.Cfg.AssetNames.BucketName(s, fileInfo.Name())
	if err != nil {
		return nil, nil, err
	}

	if bucketName == "root" && fileInfo.Name() != "/" {
		parts := strings.Split(fileInfo.Name(), "/")
		if len(parts) == 1 {
			bucketName = parts[0]
		} else {
			bucketName = parts[1]
		}
		fileInfo.FName = strings.Replace(fileInfo.Name(), "/"+bucketName, "", 1)
	}

	bucket, err := h.Cfg.Storage.GetBucket(bucketName)
	if err != nil {
		return nil, nil, err
	}

	entry.Filepath = fileInfo.Name()
	fmt.Println(entry)
	fname, err := h.Cfg.AssetNames.ObjectName(s, entry)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(bucket, fname)
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
	h.Cfg.Logger.Info(
		"listing path",
		"dir", fpath,
		"isDir", isDir,
		"recursive", recursive,
	)
	var fileList []os.FileInfo

	cleanFilename := fpath

	bucketName, err := h.Cfg.AssetNames.BucketName(s, cleanFilename)
	if err != nil {
		return fileList, err
	}
	h.Cfg.Logger.Info("fff", "name", bucketName)

	// root is a reserved bucket name so we can mount the entire object store
	if bucketName == "root" {
		if cleanFilename == "/" {
			buckets, err := h.Cfg.Storage.ListBuckets()
			if err != nil {
				return fileList, err
			}
			for _, bucket := range buckets {
				fileList = append(fileList, &utils.VirtualFile{
					FName:  bucket,
					FIsDir: true,
				})
			}
			return fileList, nil
		} else {
			parts := strings.Split(cleanFilename, "/")
			bucketName = parts[1]
			cleanFilename = strings.Replace(cleanFilename, "/"+bucketName, "", 1)
			if cleanFilename == "" && !isDir {
				info := &utils.VirtualFile{
					FName:  bucketName,
					FIsDir: true,
				}

				fileList = append(fileList, info)
				return fileList, nil
			}

			if cleanFilename == "" {
				cleanFilename = "/"
			}
		}
	}

	bucket, err := h.Cfg.Storage.GetBucket(bucketName)
	if err != nil {
		return fileList, err
	}

	fname, err := h.Cfg.AssetNames.ObjectName(s, &utils.FileEntry{Filepath: cleanFilename})
	if err != nil {
		return fileList, err
	}

	if fname == "" || fname == "." {
		name := fname
		if name == "" {
			name = "/"
		}

		info := &utils.VirtualFile{
			FName:  name,
			FIsDir: true,
		}

		fileList = append(fileList, info)
	} else {
		name := fname
		if name != "/" && isDir {
			name += "/"
		}

		foundList, err := h.Cfg.Storage.ListObjects(bucket, name, recursive)
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

	assetBucket, err := h.Cfg.AssetNames.BucketName(s, "")
	if err != nil {
		h.Cfg.Logger.Error("bucket name", "err", err)
		return err
	}
	logger := h.Cfg.Logger.With("assetBucket", assetBucket)
	bucket, err := h.Cfg.Storage.UpsertBucket(assetBucket)
	if err != nil {
		logger.Error("upsert bucket", "err", err)
		return err
	}
	setBucket(s.Context(), bucket)

	pk, _ := utils.KeyText(s)
	logger.Info(
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
	err = h.writeAsset(s, data)
	if err != nil {
		h.Cfg.Logger.Error(err.Error())
		return "", err
	}

	url, err := h.Cfg.AssetNames.PrintObjectName(s, entry, bucket.Name)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (h *UploadAssetHandler) validateAsset(data *FileData) (bool, error) {
	return true, nil
}

func (h *UploadAssetHandler) writeAsset(s ssh.Session, data *FileData) error {
	valid, err := h.validateAsset(data)
	if !valid {
		return err
	}

	objectFileName, err := h.Cfg.AssetNames.ObjectName(s, data.FileEntry)
	if err != nil {
		return err
	}
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

	_, _, err = h.Cfg.Storage.PutObject(
		data.Bucket,
		objectFileName,
		utils.NopReaderAtCloser(reader),
		data.FileEntry,
	)
	if err != nil {
		return err
	}

	return nil
}
