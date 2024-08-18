package pobj

import "github.com/picosh/send/send/utils"

type AssetNames interface {
	BucketName(user string) string
	ObjectName(entry *utils.FileEntry) string
}

type AssetNamesBasic struct{}

var _ AssetNames = &AssetNamesBasic{}
var _ AssetNames = (*AssetNamesBasic)(nil)

func (an *AssetNamesBasic) BucketName(user string) string {
	return user
}
func (an *AssetNamesBasic) ObjectName(entry *utils.FileEntry) string {
	return entry.Filepath
}

type AssetNamesForceBucket struct {
	*AssetNamesBasic
	Name string
}

func (an *AssetNamesForceBucket) BucketName(user string) string {
	return an.Name
}
