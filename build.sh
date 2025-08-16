#!/bin/bash

echo "データセット分割ツールのビルドを開始します..."

# 現在のディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 出力ディレクトリの作成
mkdir -p build

echo "macOS用バイナリをビルド中..."
GOOS=darwin GOARCH=amd64 go build -o build/dataset-splitter-mac-amd64 .
GOOS=darwin GOARCH=arm64 go build -o build/dataset-splitter-mac-arm64 .

echo "Linux用バイナリをビルド中..."
GOOS=linux GOARCH=amd64 go build -o build/dataset-splitter-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o build/dataset-splitter-linux-arm64 .

echo "Windows用バイナリをビルド中..."
GOOS=windows GOARCH=amd64 go build -o build/dataset-splitter-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -o build/dataset-splitter-windows-arm64.exe .

echo "ビルド完了！"
echo ""
echo "生成されたバイナリ:"
ls -la build/

echo ""
echo "使用方法:"
echo "  ./build/dataset-splitter-mac-amd64 [ソースディレクトリ] [出力先ディレクトリ] [教師データ比率]"
echo "  例: ./build/dataset-splitter-mac-amd64 ./鉄 ./output 0.8"
