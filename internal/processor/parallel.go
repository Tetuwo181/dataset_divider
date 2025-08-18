package processor

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"dataset-splitter/internal/utils"
)

// ProcessClassesParallel はメインクラスの並列処理
func ProcessClassesParallel(config interface{}, classDirs []string, processFunc func(string) error) error {
	if len(classDirs) == 0 {
		return nil
	}

	// 設定から並列度を取得（型アサーション）
	var maxConcurrent int
	switch cfg := config.(type) {
	case interface{ GetMaxConcurrent() int }:
		maxConcurrent = cfg.GetMaxConcurrent()
	default:
		maxConcurrent = 1
	}

	// 並列度が1の場合は順次処理
	if maxConcurrent <= 1 {
		for _, classDir := range classDirs {
			if err := processFunc(classDir); err != nil {
				log.Printf("警告: クラス %s の処理に失敗: %v", filepath.Base(classDir), err)
			}
		}
		return nil
	}

	// 並列処理
	sem := utils.NewSemaphore(maxConcurrent)
	var wg sync.WaitGroup
	errors := make(chan error, len(classDirs))

	log.Printf("並列処理を開始: %dクラスを%d並列で処理", len(classDirs), maxConcurrent)

	for _, classDir := range classDirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			if err := processFunc(dir); err != nil {
				errors <- fmt.Errorf("クラス %s の処理に失敗: %v", filepath.Base(dir), err)
			}
		}(classDir)
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
		return fmt.Errorf("一部のクラスでエラーが発生しました")
	}

	return nil
}
