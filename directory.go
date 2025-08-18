package main

import (
	"os"
	"path/filepath"
	"strings"
)

// getClassDirectories はルートディレクトリ内のクラスディレクトリを取得
func getClassDirectories(rootDir string) ([]string, error) {
	var classDirs []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// .DS_Storeなどの隠しディレクトリを除外
			if !strings.HasPrefix(entry.Name(), ".") {
				classDirs = append(classDirs, filepath.Join(rootDir, entry.Name()))
			}
		}
	}

	return classDirs, nil
}

// getSubDirectories は指定されたディレクトリ内のサブディレクトリを取得
func getSubDirectories(rootDir string) ([]string, error) {
	var subDirs []string

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// .DS_Storeなどの隠しディレクトリを除外
			if !strings.HasPrefix(entry.Name(), ".") {
				subDirs = append(subDirs, filepath.Join(rootDir, entry.Name()))
			}
		}
	}

	return subDirs, nil
}

// getImageFiles は指定されたディレクトリ内の画像ファイルを再帰的に取得
func getImageFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			// 画像ファイルの拡張子をチェック
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".bmp" {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}
