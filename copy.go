package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// copyFile は単一ファイルをコピー
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyFiles はファイル群を順次コピー
func copyFiles(destRoot, splitType, subDirName string, files []string) error {
	// 出力ディレクトリの作成
	destDir := filepath.Join(destRoot, splitType, subDirName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗: %v", err)
	}

	// ファイルのコピー
	for _, srcPath := range files {
		fileName := filepath.Base(srcPath)
		destPath := filepath.Join(destDir, fileName)

		if err := copyFile(srcPath, destPath); err != nil {
			log.Printf("警告: ファイルのコピーに失敗 %s -> %s: %v", srcPath, destPath, err)
			continue
		}
	}

	return nil
}

// copyFilesParallel はファイル群を並列コピー
func copyFilesParallel(destRoot, splitType, subDirName string, files []string, maxWorkers int) error {
	if len(files) == 0 {
		return nil
	}

	// 並列度が1の場合は順次処理
	if maxWorkers <= 1 {
		return copyFiles(destRoot, splitType, subDirName, files)
	}

	// 出力ディレクトリの作成
	destDir := filepath.Join(destRoot, splitType, subDirName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗: %v", err)
	}

	// 並列コピー処理
	sem := NewSemaphore(maxWorkers)
	var wg sync.WaitGroup
	errors := make(chan error, len(files))

	for _, srcPath := range files {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			fileName := filepath.Base(src)
			destPath := filepath.Join(destDir, fileName)

			if err := copyFile(src, destPath); err != nil {
				errors <- fmt.Errorf("ファイルのコピーに失敗 %s -> %s: %v", src, destPath, err)
			}
		}(srcPath)
	}

	wg.Wait()
	close(errors)

	// エラーの確認
	var hasErrors bool
	for err := range errors {
		log.Printf("警告: %v", err)
		hasErrors = true
	}

	if hasErrors {
		return fmt.Errorf("一部のファイルのコピーに失敗しました")
	}

	return nil
}
