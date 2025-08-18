package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
)

// Semaphore は並列処理用のセマフォ
type Semaphore struct {
	sem chan struct{}
}

// NewSemaphore は新しいセマフォを作成
func NewSemaphore(max int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, max),
	}
}

// Acquire はセマフォを取得
func (s *Semaphore) Acquire() {
	s.sem <- struct{}{}
}

// Release はセマフォを解放
func (s *Semaphore) Release() {
	<-s.sem
}

// processClassesParallel はメインクラスの並列処理
func processClassesParallel(config *Config, classDirs []string) error {
	if len(classDirs) == 0 {
		return nil
	}

	// 並列度が1の場合は順次処理
	if config.MaxConcurrent <= 1 {
		for _, classDir := range classDirs {
			if err := processClassDirectory(config, classDir); err != nil {
				log.Printf("警告: クラス %s の処理に失敗: %v", filepath.Base(classDir), err)
			}
		}
		return nil
	}

	// 並列処理
	sem := NewSemaphore(config.MaxConcurrent)
	var wg sync.WaitGroup
	errors := make(chan error, len(classDirs))

	log.Printf("並列処理を開始: %dクラスを%d並列で処理", len(classDirs), config.MaxConcurrent)

	for _, classDir := range classDirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			if err := processClassDirectory(config, dir); err != nil {
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
