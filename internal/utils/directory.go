package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetClassDirectories はルートディレクトリ内のクラスディレクトリを取得
func GetClassDirectories(rootDir string) ([]string, error) {
	var classDirs []string
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !strings.HasPrefix(entry.Name(), ".") {
				classDirs = append(classDirs, filepath.Join(rootDir, entry.Name()))
			}
		}
	}
	return classDirs, nil
}

// GetSubDirectories は指定されたディレクトリ内のサブディレクトリを取得
func GetSubDirectories(rootDir string) ([]string, error) {
	var subDirs []string
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if !strings.HasPrefix(entry.Name(), ".") {
				subDirs = append(subDirs, filepath.Join(rootDir, entry.Name()))
			}
		}
	}
	return subDirs, nil
}

// GetImageFiles は指定されたディレクトリ内の画像ファイルを再帰的に取得
func GetImageFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".bmp" {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// GetClassName はディレクトリパスからクラス名を取得
func GetClassName(dirPath string) string {
	return filepath.Base(dirPath)
}
