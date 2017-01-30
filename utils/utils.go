package utils

import (
	"os"
	"path/filepath"
)

// FileExist 文件是否存在
func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// GetImageRePath 获取文件相对路径
func GetImageRePath(name string) string {
	return filepath.Join(name[0:2], name)
}
