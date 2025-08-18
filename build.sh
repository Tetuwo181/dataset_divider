#!/bin/bash

echo "🚀 Dataset Splitter ビルド開始..."

# コマンドライン版のビルド
echo "📦 コマンドライン版をビルド中..."
go build -o dataset-splitter .
if [ $? -eq 0 ]; then
    echo "✅ コマンドライン版のビルド成功: dataset-splitter"
else
    echo "❌ コマンドライン版のビルド失敗"
    exit 1
fi

# バイナリサイズの表示
echo ""
echo "📊 ビルド結果:"
echo "  コマンドライン版: $(ls -lh dataset-splitter | awk '{print $5}')"
echo ""
echo "🎉 ビルド完了！"
echo ""
echo "使用方法:"
echo "  コマンドライン版: ./dataset-splitter -help"
echo ""
echo "機能:"
echo "  - 多クラス分類データセット分割"
echo "  - 二値分類モード"
echo "  - tar出力"
echo "  - 並列処理"
echo "  - 最小ファイル数フィルタリング"
