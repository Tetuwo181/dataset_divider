package main

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
)

// processBinaryClassification は二値分類処理
func processBinaryClassification(config *Config, classDirs []string) error {
	// positiveクラスのデータを収集
	var positiveFiles []string
	var allOtherFiles []string

	log.Printf("positiveクラス '%s' のデータを収集中...", config.PositiveClass)
	log.Printf("二値分類モード: 最小ファイル数制限を無効化（全サブクラスからデータを取得）")

	// 全クラスディレクトリを走査してpositiveクラスとその他のデータを分類
	for _, classDir := range classDirs {
		className := filepath.Base(classDir)
		log.Printf("クラス '%s' を処理中...", className)

		// サブディレクトリを取得
		subDirs, err := getSubDirectories(classDir)
		if err != nil {
			log.Printf("警告: クラス %s のサブディレクトリ取得に失敗: %v", className, err)
			continue
		}

		// 各サブクラスから均等にデータを取得
		var classFiles []string
		for _, subDir := range subDirs {
			subDirName := filepath.Base(subDir)
			log.Printf("  サブディレクトリ '%s' を処理中...", subDirName)

			// 画像ファイルを取得
			files, err := getImageFiles(subDir)
			if err != nil {
				log.Printf("    警告: ファイル一覧の取得に失敗: %v", err)
				continue
			}

			if len(files) == 0 {
				log.Printf("    警告: サブディレクトリ '%s' に画像ファイルがありません", subDirName)
				continue
			}

			log.Printf("    ファイル数: %d", len(files))

			// 二値分類モードでは最小ファイル数制限を無効化
			// すべてのサブクラスからデータを取得
			if len(files) > 0 {
				// サブクラス内でシャッフル
				rand.Shuffle(len(files), func(i, j int) {
					files[i], files[j] = files[j], files[i]
				})

				// サブクラスからデータを追加
				classFiles = append(classFiles, files...)
				log.Printf("    追加: %d件", len(files))
			}
		}

		if len(classFiles) == 0 {
			log.Printf("警告: クラス '%s' に有効な画像ファイルがありません", className)
			continue
		}

		log.Printf("  クラス全体のファイル数: %d", len(classFiles))

		// positiveクラスかどうかで分類
		if className == config.PositiveClass {
			positiveFiles = append(positiveFiles, classFiles...)
			log.Printf("  positiveクラスとして追加: %d件", len(classFiles))
		} else {
			allOtherFiles = append(allOtherFiles, classFiles...)
			log.Printf("  negativeクラスとして追加: %d件", len(classFiles))
		}
	}

	// データ数の確認
	log.Printf("positiveクラス: %d件", len(positiveFiles))
	log.Printf("negativeクラス: %d件", len(allOtherFiles))

	if len(positiveFiles) == 0 {
		return fmt.Errorf("positiveクラス '%s' のデータが見つかりません", config.PositiveClass)
	}

	// 少ない方のデータ数を基準に設定
	targetCount := len(positiveFiles)
	if len(allOtherFiles) < targetCount {
		targetCount = len(allOtherFiles)
	}

	log.Printf("均等化後のデータ数: %d件 (positive: %d, negative: %d)", targetCount*2, targetCount, targetCount)

	// データをシャッフル
	rand.Shuffle(len(positiveFiles), func(i, j int) {
		positiveFiles[i], positiveFiles[j] = positiveFiles[j], positiveFiles[i]
	})
	rand.Shuffle(len(allOtherFiles), func(i, j int) {
		allOtherFiles[i], allOtherFiles[j] = allOtherFiles[j], allOtherFiles[i]
	})

	// 分割点の計算
	trainingCount := int(float64(targetCount) * config.TrainingRatio)

	// 教師データと検証データに分割
	positiveTraining := positiveFiles[:trainingCount]
	positiveValidation := positiveFiles[trainingCount:targetCount]
	negativeTraining := allOtherFiles[:trainingCount]
	negativeValidation := allOtherFiles[trainingCount:targetCount]

	// ディレクトリの作成とファイルのコピー
	log.Printf("positive/negativeデータのコピーを開始...")

	// positiveクラスのコピー
	if err := copyFilesParallel(config.DestDir, "train", "positive", positiveTraining, config.MaxCopyWorkers); err != nil {
		return fmt.Errorf("positive教師データのコピーに失敗: %v", err)
	}
	if err := copyFilesParallel(config.DestDir, "validation", "positive", positiveValidation, config.MaxCopyWorkers); err != nil {
		return fmt.Errorf("positive検証データのコピーに失敗: %v", err)
	}

	// negativeクラスのコピー
	if err := copyFilesParallel(config.DestDir, "train", "negative", negativeTraining, config.MaxCopyWorkers); err != nil {
		return fmt.Errorf("negative教師データのコピーに失敗: %v", err)
	}
	if err := copyFilesParallel(config.DestDir, "validation", "negative", negativeValidation, config.MaxCopyWorkers); err != nil {
		return fmt.Errorf("negative検証データのコピーに失敗: %v", err)
	}

	log.Printf("二値分類データセットの作成が完了しました！")
	log.Printf("  教師データ: positive %d件, negative %d件", len(positiveTraining), len(negativeTraining))
	log.Printf("  検証データ: positive %d件, negative %d件", len(positiveValidation), len(negativeValidation))

	return nil
}
