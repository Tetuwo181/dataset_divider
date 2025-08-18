package config

import (
	"fmt"
	"runtime"
)

// Config は設定情報を保持
type Config struct {
	SourceDir      string  // ソースディレクトリ
	DestDir        string  // 出力先ディレクトリ
	TrainingRatio  float64 // 教師データ比率
	MinFileCount   int     // 最小ファイル数
	TarOutput      bool    // tar出力フラグ
	MaxConcurrent  int     // 最大並列度
	MaxCopyWorkers int     // 最大コピーワーカー数
	BinaryMode     bool    // 二値分類モード
	PositiveClass  string  // positiveクラス名
}

// NewDefaultConfig はデフォルト設定を返す
func NewDefaultConfig() *Config {
	return &Config{
		TrainingRatio:  0.7,
		MinFileCount:   50,
		TarOutput:      false,
		MaxConcurrent:  runtime.NumCPU() / 2,
		MaxCopyWorkers: runtime.NumCPU(),
		BinaryMode:     false,
		PositiveClass:  "",
	}
}

// Validate は設定の妥当性をチェック
func (c *Config) Validate() error {
	if c.SourceDir == "" {
		return fmt.Errorf("ソースディレクトリが指定されていません")
	}
	if c.DestDir == "" {
		return fmt.Errorf("出力先ディレクトリが指定されていません")
	}
	if c.TrainingRatio <= 0.0 || c.TrainingRatio >= 1.0 {
		return fmt.Errorf("教師データ比率は0.0より大きく1.0より小さい値である必要があります")
	}
	if c.MinFileCount < 1 {
		return fmt.Errorf("最小ファイル数は1以上である必要があります")
	}
	if c.MaxConcurrent < 1 {
		return fmt.Errorf("最大並列度は1以上である必要があります")
	}
	if c.MaxCopyWorkers < 1 {
		return fmt.Errorf("最大コピーワーカー数は1以上である必要があります")
	}
	if c.BinaryMode && c.PositiveClass == "" {
		return fmt.Errorf("二値分類モードではpositiveクラスを指定する必要があります")
	}
	return nil
}

// GetValidationRatio は検証データ比率を返す
func (c *Config) GetValidationRatio() float64 {
	return 1.0 - c.TrainingRatio
}

// GetMaxConcurrent は最大並列度を返す
func (c *Config) GetMaxConcurrent() int {
	return c.MaxConcurrent
}

// GetMaxCopyWorkers は最大コピーワーカー数を返す
func (c *Config) GetMaxCopyWorkers() int {
	return c.MaxCopyWorkers
}
