package utils

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
