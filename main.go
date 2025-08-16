package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	SourceDir       string
	DestDir         string
	TrainingRatio   float64
	ValidationRatio float64
	MinFileCount    int
}

func main() {
	// コマンドライン引数の解析
	config := parseFlags()

	// 引数の検証
	if err := validateConfig(config); err != nil {
		log.Fatalf("設定エラー: %v", err)
	}

	// ソースディレクトリの存在確認
	if _, err := os.Stat(config.SourceDir); os.IsNotExist(err) {
		log.Fatalf("ソースディレクトリが存在しません: %s", config.SourceDir)
	}

	// ランダムシードの設定
	rand.Seed(time.Now().UnixNano())

	// 処理開始
	log.Printf("データセット分割を開始します...")
	log.Printf("ソース: %s", config.SourceDir)
	log.Printf("出力先: %s", config.DestDir)
	log.Printf("教師データ比率: %.2f%%", config.TrainingRatio*100)
	log.Printf("検証データ比率: %.2f%%", config.ValidationRatio*100)
	log.Printf("最小ファイル数: %d", config.MinFileCount)

	// クラスディレクトリの取得
	classDirs, err := getClassDirectories(config.SourceDir)
	if err != nil {
		log.Fatalf("クラスディレクトリの取得に失敗: %v", err)
	}

	log.Printf("検出されたクラス数: %d", len(classDirs))

	// 各クラスディレクトリの処理
	for _, classDir := range classDirs {
		if err := processClassDirectory(config, classDir); err != nil {
			log.Printf("警告: クラス %s の処理に失敗: %v", classDir, err)
			continue
		}
	}

	log.Printf("データセット分割が完了しました！")
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.SourceDir, "source", "", "ソースディレクトリのパス")
	flag.StringVar(&config.DestDir, "dest", "", "出力先ディレクトリのパス")
	flag.Float64Var(&config.TrainingRatio, "ratio", 0.8, "教師データの比率 (0.0-1.0)")
	flag.IntVar(&config.MinFileCount, "min-files", 50, "コピーする最小ファイル数")

	flag.Parse()

	// 位置引数もサポート
	args := flag.Args()
	if len(args) >= 1 && config.SourceDir == "" {
		config.SourceDir = args[0]
	}
	if len(args) >= 2 && config.DestDir == "" {
		config.DestDir = args[1]
	}
	if len(args) >= 3 && config.TrainingRatio == 0.8 {
		if ratio, err := parseRatio(args[2]); err == nil {
			config.TrainingRatio = ratio
		}
	}

	config.ValidationRatio = 1.0 - config.TrainingRatio

	return config
}

func parseRatio(ratioStr string) (float64, error) {
	var ratio float64
	_, err := fmt.Sscanf(ratioStr, "%f", &ratio)
	if err != nil {
		return 0, err
	}

	// パーセンテージ表記もサポート
	if strings.HasSuffix(ratioStr, "%") {
		ratio = ratio / 100.0
	}

	return ratio, nil
}

func validateConfig(config *Config) error {
	if config.SourceDir == "" {
		return fmt.Errorf("ソースディレクトリが指定されていません")
	}

	if config.DestDir == "" {
		return fmt.Errorf("出力先ディレクトリが指定されていません")
	}

	if config.TrainingRatio < 0.0 || config.TrainingRatio > 1.0 {
		return fmt.Errorf("教師データ比率は0.0から1.0の間である必要があります")
	}

	return nil
}

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

func processClassDirectory(config *Config, classDir string) error {
	className := filepath.Base(classDir)
	log.Printf("クラス '%s' を処理中...", className)

	// クラスディレクトリ直下のサブディレクトリを取得
	subDirs, err := getSubDirectories(classDir)
	if err != nil {
		return fmt.Errorf("サブディレクトリの取得に失敗: %v", err)
	}

	if len(subDirs) == 0 {
		log.Printf("警告: クラス '%s' にサブディレクトリがありません", className)
		return nil
	}

	log.Printf("  サブディレクトリ数: %d", len(subDirs))

	// 各サブディレクトリの処理
	for _, subDir := range subDirs {
		subDirName := filepath.Base(subDir)
		log.Printf("    サブディレクトリ '%s' を処理中...", subDirName)

		// サブディレクトリ内の画像ファイルを取得
		files, err := getImageFiles(subDir)
		if err != nil {
			log.Printf("      警告: ファイル一覧の取得に失敗: %v", err)
			continue
		}

		if len(files) == 0 {
			log.Printf("      警告: サブディレクトリ '%s' に画像ファイルがありません", subDirName)
			continue
		}

		log.Printf("      ファイル数: %d", len(files))

		// 最小ファイル数チェック
		if len(files) < config.MinFileCount {
			log.Printf("      スキップ: ファイル数が%d未満のため (%d < %d)", config.MinFileCount, len(files), config.MinFileCount)
			continue
		}

		// ファイルをシャッフル
		shuffledFiles := make([]string, len(files))
		copy(shuffledFiles, files)
		rand.Shuffle(len(shuffledFiles), func(i, j int) {
			shuffledFiles[i], shuffledFiles[j] = shuffledFiles[j], shuffledFiles[i]
		})

		// 分割点の計算
		trainingCount := int(float64(len(shuffledFiles)) * config.TrainingRatio)

		// 教師データと検証データに分割
		trainingFiles := shuffledFiles[:trainingCount]
		validationFiles := shuffledFiles[trainingCount:]

		// ディレクトリの作成とファイルのコピー
		if err := copyFiles(config.DestDir, "train", subDirName, trainingFiles); err != nil {
			log.Printf("      警告: 教師データのコピーに失敗: %v", err)
			continue
		}

		if err := copyFiles(config.DestDir, "validation", subDirName, validationFiles); err != nil {
			log.Printf("      警告: 検証データのコピーに失敗: %v", err)
			continue
		}

		log.Printf("      教師データ: %d件, 検証データ: %d件", len(trainingFiles), len(validationFiles))
	}

	return nil
}

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
