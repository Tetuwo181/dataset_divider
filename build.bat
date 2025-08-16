@echo off
echo データセット分割ツールのビルドを開始します...

REM 出力ディレクトリの作成
if not exist build mkdir build

echo macOS用バイナリをビルド中...
set GOOS=darwin
set GOARCH=amd64
go build -o build\dataset-splitter-mac-amd64 .

set GOOS=darwin
set GOARCH=arm64
go build -o build\dataset-splitter-mac-arm64 .

echo Linux用バイナリをビルド中...
set GOOS=linux
set GOARCH=amd64
go build -o build\dataset-splitter-linux-amd64 .

set GOOS=linux
set GOARCH=arm64
go build -o build\dataset-splitter-linux-arm64 .

echo Windows用バイナリをビルド中...
set GOOS=windows
set GOARCH=amd64
go build -o build\dataset-splitter-windows-amd64.exe .

set GOOS=windows
set GOARCH=arm64
go build -o build\dataset-splitter-windows-arm64.exe .

echo ビルド完了！
echo.
echo 生成されたバイナリ:
dir build

echo.
echo 使用方法:
echo   build\dataset-splitter-windows-amd64.exe [ソースディレクトリ] [出力先ディレクトリ] [教師データ比率]
echo   例: build\dataset-splitter-windows-amd64.exe .\鉄 .\output 0.8

pause
