package objx

import "github.com/picosh/send/send/utils"

type AssetNames interface {
	BucketName(user string) string
	FileName(entry *utils.FileEntry) string
}

type AssetNamesBasic struct {}
var _ AssetNames = &AssetNamesBasic{}
var _ AssetNames = (*AssetNamesBasic)(nil)

func (an *AssetNamesBasic) BucketName(user string) string {
	return user
}
func (an *AssetNamesBasic) FileName(entry *utils.FileEntry) string {
	return entry.Filepath
}
