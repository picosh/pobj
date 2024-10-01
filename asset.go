package pobj

import (
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/picosh/send/send/utils"
)

type AssetNames interface {
	BucketName(sesh ssh.Session, fpath string) (string, error)
	ObjectName(sesh ssh.Session, entry *utils.FileEntry) (string, error)
	PrintObjectName(sesh ssh.Session, entry *utils.FileEntry, bucketName string) (string, error)
}

type AssetNamesBasic struct{}

var _ AssetNames = &AssetNamesBasic{}
var _ AssetNames = (*AssetNamesBasic)(nil)

func (an *AssetNamesBasic) BucketName(sesh ssh.Session, fpath string) (string, error) {
	return sesh.User(), nil
}
func (an *AssetNamesBasic) ObjectName(sesh ssh.Session, entry *utils.FileEntry) (string, error) {
	return entry.Filepath, nil
}
func (an *AssetNamesBasic) PrintObjectName(sesh ssh.Session, entry *utils.FileEntry, bucketName string) (string, error) {
	objectName, err := an.ObjectName(sesh, entry)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", bucketName, objectName), nil
}

type AssetNamesForceBucket struct {
	*AssetNamesBasic
	Name string
}

var _ AssetNames = &AssetNamesForceBucket{}
var _ AssetNames = (*AssetNamesForceBucket)(nil)

func (an *AssetNamesForceBucket) BucketName(sesh ssh.Session, fpath string) (string, error) {
	return an.Name, nil
}
