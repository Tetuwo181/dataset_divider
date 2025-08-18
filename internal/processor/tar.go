package processor

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// CreateTarArchive はtarアーカイブを作成
func CreateTarArchive(sourceDir string) error {
	// tarファイル名を生成（ディレクトリ名 + .tar）
	dirName := filepath.Base(sourceDir)
	tarFileName := dirName + ".tar"

	log.Printf("tarファイルの作成を開始: %s", tarFileName)

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
			if err == io.EOF {
				break
			}
			return total, err
		}
	}

	return total, nil
}
