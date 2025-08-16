# データセット分割ツール (Dataset Splitter)

機械学習のクラス分けに使用する画像データを、指定した比率で教師データと検証データに分割するツールです。

## 機能

- ディレクトリ構造を保持したままデータを分割
- 各クラスディレクトリからランダムにファイルを選択
- 教師データと検証データの比率を指定可能
- 画像ファイル（jpg, jpeg, png, gif, bmp）を自動検出
- 指定した枚数以下のディレクトリは自動的にスキップ
- クロスプラットフォーム対応（Windows, macOS, Linux）

## 使用方法

### コマンドラインオプション

- `-source`: ソースディレクトリのパス
- `-dest`: 出力先ディレクトリのパス  
- `-ratio`: 教師データの比率 (0.0-1.0、デフォルト: 0.8)
- `-min-files`: コピーする最小ファイル数 (デフォルト: 50)

### 基本的な使用方法

```bash
# 位置引数を使用
./dataset-splitter [ソースディレクトリ] [出力先ディレクトリ] [教師データ比率]

# フラグを使用
./dataset-splitter -source [ソースディレクトリ] -dest [出力先ディレクトリ] -ratio [教師データ比率] -min-files [最小ファイル数]
```

### 使用例

```bash
# 鉄道画像データを80%の教師データ、20%の検証データに分割
./dataset-splitter ./鉄 ./output 0.8

# モノレール画像データを70%の教師データ、30%の検証データに分割
./dataset-splitter ./懸垂式モノレール ./monorail_output 0.7

# パーセンテージ表記も使用可能
./dataset-splitter ./跨座式モノレール ./monorail_output 75%

# 最小ファイル数を指定（30枚以下のディレクトリはスキップ）
./dataset-splitter -source ./鉄道画像 -dest ./output -min-files 30
```

## 出力構造

ツールは以下のような構造でデータを出力します：

```
出力先ディレクトリ/
├── train/                    # 教師データ
│   ├── 103系/
│   ├── 115系/
│   ├── 205系/
│   └── ...
└── validation/               # 検証データ
    ├── 103系/
    ├── 115系/
    ├── 205系/
    └── ...
```

## ビルド方法

### macOS/Linux

```bash
# ビルドスクリプトを実行
chmod +x build.sh
./build.sh
```

### Windows

```cmd
# バッチファイルを実行
build.bat
```

### 手動ビルド

```bash
# 現在のプラットフォーム用
go build -o dataset-splitter .

# 特定のプラットフォーム用
GOOS=darwin GOARCH=amd64 go build -o dataset-splitter-mac-amd64 .
GOOS=windows GOARCH=amd64 go build -o dataset-splitter-windows-amd64.exe .
```

## 生成されるバイナリ

ビルド後、以下のバイナリが生成されます：

- `dataset-splitter-mac-amd64` - macOS (Intel)
- `dataset-splitter-mac-arm64` - macOS (Apple Silicon)
- `dataset-splitter-linux-amd64` - Linux (Intel)
- `dataset-splitter-linux-arm64` - Linux (ARM)
- `dataset-splitter-windows-amd64.exe` - Windows (Intel)
- `dataset-splitter-windows-arm64.exe` - Windows (ARM)

## 注意事項

- ソースディレクトリ内の各サブディレクトリが1つのクラスとして扱われます
- 画像ファイル以外のファイルは無視されます
- 指定した最小ファイル数未満のディレクトリは自動的にスキップされます（デフォルト: 50枚）
- 既存の出力ディレクトリがある場合、上書きされます
- ファイルのコピーに失敗した場合は警告が表示されますが、処理は継続されます

## ライセンス

このツールはMITライセンスの下で公開されています。
