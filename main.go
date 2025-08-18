package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	SourceDir       string
	DestDir         string
	TrainingRatio   float64
	ValidationRatio float64
	MinFileCount    int
	TarOutput       bool
	MaxConcurrent   int
	MaxCopyWorkers  int
	BinaryMode      bool
	PositiveClass   string
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
	log.Printf("tar出力: %t", config.TarOutput)
	log.Printf("並列処理: %dクラス, %dワーカー", config.MaxConcurrent, config.MaxCopyWorkers)
	if config.BinaryMode {
		log.Printf("二値分類モード: positiveクラス '%s'", config.PositiveClass)
	}

	// クラスディレクトリの取得
	classDirs, err := getClassDirectories(config.SourceDir)
	if err != nil {
		log.Fatalf("クラスディレクトリの取得に失敗: %v", err)
	}

	log.Printf("検出されたクラス数: %d", len(classDirs))

	// 二値分類モードか通常モードかを判定
	if config.BinaryMode {
		log.Printf("二値分類モード: positiveクラス '%s'", config.PositiveClass)
		if err := processBinaryClassification(config, classDirs); err != nil {
			log.Fatalf("二値分類処理中にエラーが発生しました: %v", err)
		}
	} else {
		// 通常の多クラス分類処理
		if err := processClassesParallel(config, classDirs); err != nil {
			log.Fatalf("クラスディレクトリの処理中にエラーが発生しました: %v", err)
		}
	}

	log.Printf("データセット分割が完了しました！")

	// tarファイル作成オプションが有効な場合
	if config.TarOutput {
		log.Printf("tarファイルの作成を開始します...")
		if err := createTarArchive(config.DestDir); err != nil {
			log.Printf("警告: tarファイルの作成に失敗: %v", err)
		} else {
			log.Printf("tarファイルの作成が完了しました！")
		}
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.SourceDir, "source", "", "ソースディレクトリのパス")
	flag.StringVar(&config.DestDir, "dest", "", "出力先ディレクトリのパス")
	flag.Float64Var(&config.TrainingRatio, "ratio", 0.8, "教師データの比率 (0.0-1.0)")
	flag.IntVar(&config.MinFileCount, "min-files", 50, "コピーする最小ファイル数")
	flag.BoolVar(&config.TarOutput, "tar", false, "出力をtarファイルに圧縮")
	flag.IntVar(&config.MaxConcurrent, "max-concurrent", 0, "同時処理するクラス数 (0=自動設定)")
	flag.IntVar(&config.MaxCopyWorkers, "copy-workers", 0, "ファイルコピーの並列数 (0=自動設定)")
	flag.BoolVar(&config.BinaryMode, "binary", false, "二値分類モード（positive/negative）")
	flag.StringVar(&config.PositiveClass, "positive", "", "positiveクラスのサブディレクトリ名（二値分類モード時）")

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

	// デフォルト値の設定
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = runtime.NumCPU() / 2
		if config.MaxConcurrent < 1 {
			config.MaxConcurrent = 1
		}
	}
	if config.MaxCopyWorkers == 0 {
		config.MaxCopyWorkers = runtime.NumCPU()
		if config.MaxCopyWorkers < 1 {
			config.MaxCopyWorkers = 1
		}
	}

	// 二値分類モードの検証
	if config.BinaryMode && config.PositiveClass == "" {
		log.Fatal("二値分類モードでは -positive オプションが必須です")
	}

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
		if err := copyFilesParallel(config.DestDir, "train", subDirName, trainingFiles, config.MaxCopyWorkers); err != nil {
			log.Printf("      警告: 教師データのコピーに失敗: %v", err)
			continue
		}

		if err := copyFilesParallel(config.DestDir, "validation", subDirName, validationFiles, config.MaxCopyWorkers); err != nil {
			log.Printf("      警告: 検証データのコピーに失敗: %v", err)
			continue
		}

		log.Printf("      教師データ: %d件, 検証データ: %d件", len(trainingFiles), len(validationFiles))
	}

	return nil
}

func createTarArchive(sourceDir string) error {
	// tarファイル名を生成（ディレクトリ名 + .tar）
	dirName := filepath.Base(sourceDir)
	tarFileName := dirName + ".tar"

	// tarファイルを作成
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		return fmt.Errorf("tarファイルの作成に失敗: %v", err)
	}
	defer tarFile.Close()

	// tarライターを作成
	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	// ディレクトリ内のファイルを再帰的にtarに追加
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ソースディレクトリ自体はスキップ
		if path == sourceDir {
			return nil
		}

		// 相対パスを計算
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// tarヘッダーを作成
		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath

		// ヘッダーをtarに書き込み
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// ディレクトリの場合はファイル内容を書き込まない
		if info.IsDir() {
			return nil
		}

		// ファイルの内容をtarに書き込み
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = copyFileContent(file, tarWriter)
		return err
	})

	if err != nil {
		return fmt.Errorf("ファイルのtar化に失敗: %v", err)
	}

	log.Printf("tarファイルが作成されました: %s", tarFileName)
	return nil
}

// copyFileContent はファイルの内容をtarライターにコピー
func copyFileContent(file *os.File, tarWriter *tar.Writer) (int64, error) {
	buf := make([]byte, 32*1024) // 32KBバッファ
	var total int64

	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, writeErr := tarWriter.Write(buf[:n]); writeErr != nil {
				return total, writeErr
			}
			total += int64(n)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return total, err
		}
	}

	return total, nil
}
