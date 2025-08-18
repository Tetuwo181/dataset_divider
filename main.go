package main

import (
	"flag"
	"fmt"
	"log"

	"dataset-splitter/internal/config"
	"dataset-splitter/internal/processor"
	"dataset-splitter/internal/utils"
)

func main() {
	// フラグの解析
	config := parseFlags()

	// 設定の検証
	if err := config.Validate(); err != nil {
		log.Fatalf("設定エラー: %v", err)
	}

	// ログ出力
	log.Printf("データセット分割を開始します...")
	log.Printf("ソース: %s", config.SourceDir)
	log.Printf("出力先: %s", config.DestDir)
	log.Printf("教師データ比率: %.2f%%", config.TrainingRatio*100)
	log.Printf("検証データ比率: %.2f%%", config.GetValidationRatio()*100)
	log.Printf("最小ファイル数: %d", config.MinFileCount)
	log.Printf("tar出力: %t", config.TarOutput)
	log.Printf("並列処理: %dクラス, %dワーカー", config.MaxConcurrent, config.MaxCopyWorkers)

	if config.BinaryMode {
		log.Printf("二値分類モード: positiveクラス '%s'", config.PositiveClass)
	}

	// クラスディレクトリの取得
	classDirs, err := utils.GetClassDirectories(config.SourceDir)
	if err != nil {
		log.Fatalf("クラスディレクトリの取得に失敗: %v", err)
	}

	log.Printf("検出されたクラス数: %d", len(classDirs))

	// 処理の実行
	if config.BinaryMode {
		if err := processBinaryClassification(config, classDirs); err != nil {
			log.Fatalf("二値分類処理に失敗: %v", err)
		}
	} else {
		if err := processClassesParallel(config, classDirs); err != nil {
			log.Fatalf("並列処理に失敗: %v", err)
		}
	}

	// tar出力
	if config.TarOutput {
		if err := createTarArchive(config.DestDir); err != nil {
			log.Printf("警告: tar出力に失敗: %v", err)
		}
	}

	log.Printf("データセット分割が完了しました！")
}

// parseFlags はコマンドライン引数を解析
func parseFlags() *config.Config {
	cfg := config.NewDefaultConfig()

	flag.StringVar(&cfg.SourceDir, "source", "", "ソースディレクトリ")
	flag.StringVar(&cfg.DestDir, "dest", "", "出力先ディレクトリ")
	flag.Float64Var(&cfg.TrainingRatio, "ratio", cfg.TrainingRatio, "教師データ比率 (0.0-1.0)")
	flag.IntVar(&cfg.MinFileCount, "min-files", cfg.MinFileCount, "最小ファイル数")
	flag.BoolVar(&cfg.TarOutput, "tar", cfg.TarOutput, "tar出力フラグ")
	flag.IntVar(&cfg.MaxConcurrent, "max-concurrent", cfg.MaxConcurrent, "最大並列度")
	flag.IntVar(&cfg.MaxCopyWorkers, "copy-workers", cfg.MaxCopyWorkers, "最大コピーワーカー数")
	flag.BoolVar(&cfg.BinaryMode, "binary", cfg.BinaryMode, "二値分類モード")
	flag.StringVar(&cfg.PositiveClass, "positive", cfg.PositiveClass, "positiveクラス名")

	flag.Parse()

	return cfg
}

// processClassesParallel は並列処理を実行
func processClassesParallel(config *config.Config, classDirs []string) error {
	return processor.ProcessClassesParallel(config, classDirs, func(classDir string) error {
		return processClassDirectory(config, classDir)
	})
}

// processClassDirectory は個別クラスディレクトリを処理
func processClassDirectory(config *config.Config, classDir string) error {
	className := utils.GetClassName(classDir)
	log.Printf("クラス '%s' を処理中...", className)

	// サブディレクトリの取得
	subDirs, err := utils.GetSubDirectories(classDir)
	if err != nil {
		return fmt.Errorf("サブディレクトリの取得に失敗: %v", err)
	}

	// 各サブディレクトリを処理
	for _, subDir := range subDirs {
		subDirName := utils.GetClassName(subDir)
		log.Printf("  サブディレクトリ '%s' を処理中...", subDirName)

		// 画像ファイルの取得
		files, err := utils.GetImageFiles(subDir)
		if err != nil {
			log.Printf("    警告: ファイル一覧の取得に失敗: %v", err)
			continue
		}

		if len(files) == 0 {
			log.Printf("    警告: サブディレクトリ '%s' に画像ファイルがありません", subDirName)
			continue
		}

		log.Printf("    ファイル数: %d", len(files))

		// 最小ファイル数チェック
		if len(files) < config.MinFileCount {
			log.Printf("    スキップ: ファイル数が%d未満のため (%d < %d)", config.MinFileCount, len(files), config.MinFileCount)
			continue
		}

		// ファイルの分割
		trainingCount := int(float64(len(files)) * config.TrainingRatio)
		trainingFiles := files[:trainingCount]
		validationFiles := files[trainingCount:]

		// ファイルのコピー
		if err := processor.CopyFilesParallel(config.DestDir, "train", subDirName, trainingFiles, config.MaxCopyWorkers); err != nil {
			log.Printf("    警告: 教師データのコピーに失敗: %v", err)
		}

		if err := processor.CopyFilesParallel(config.DestDir, "validation", subDirName, validationFiles, config.MaxCopyWorkers); err != nil {
			log.Printf("    警告: 検証データのコピーに失敗: %v", err)
		}

		log.Printf("    完了: 教師データ %d件, 検証データ %d件", len(trainingFiles), len(validationFiles))
	}

	return nil
}

// processBinaryClassification は二値分類処理
func processBinaryClassification(config *config.Config, classDirs []string) error {
	return processor.ProcessBinaryClassification(config, classDirs)
}

// createTarArchive はtarアーカイブを作成
func createTarArchive(destDir string) error {
	return processor.CreateTarArchive(destDir)
}
