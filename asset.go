package pobj

import (
	"github.com/charmbracelet/ssh"
	"github.com/picosh/send/send/utils"
)

type AssetNames interface {
	BucketName(sesh ssh.Session) (string, error)
	ObjectName(sesh ssh.Session, entry *utils.FileEntry) (string, error)
}

type AssetNamesBasic struct{}

var _ AssetNames = &AssetNamesBasic{}
var _ AssetNames = (*AssetNamesBasic)(nil)

func (an *AssetNamesBasic) BucketName(sesh ssh.Session) (string, error) {
	return sesh.User(), nil
}
func (an *AssetNamesBasic) ObjectName(sesh ssh.Session, entry *utils.FileEntry) (string, error) {
	return entry.Filepath, nil
}

type AssetNamesForceBucket struct {
	*AssetNamesBasic
	Name string
}

var _ AssetNames = &AssetNamesForceBucket{}
var _ AssetNames = (*AssetNamesForceBucket)(nil)

func (an *AssetNamesForceBucket) BucketName(sesh ssh.Session) (string, error) {
	return an.Name, nil
}
