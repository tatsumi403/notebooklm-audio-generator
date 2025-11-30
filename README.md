# NotebookLM Audio Generator

URLを追加するだけでNotebookLMに自動的にソースとして登録し、音声解説を生成します。

## 🚀 実装バージョン

このプロジェクトはPythonとGo言語の2つのバージョンを提供しています:

### Python版 (scripts/add_to_notebooklm.py)
- Selenium + webdriver-managerを使用
- 従来の実装、安定動作確認済み

### Go版 (scripts/add_to_notebooklm.go) ⭐ 推奨
- chromedpを使用した純粋なGo実装
- **メリット:**
  - バイナリ1つで動作、Python環境不要
  - 高速起動・低メモリ使用量
  - 型安全性によるバグ削減
  - 並行処理が容易

## 使い方

1. `urls.txt` に追加したいURLを1行ずつ記載
2. Git commit & push
3. GitHub Actionsが自動実行され、NotebookLMに追加・音声生成

## セットアップ

### 1. OAuth認証情報の取得

1. [Google Cloud Console](https://console.cloud.google.com/) にアクセス
2. 新しいプロジェクトを作成
3. 「APIとサービス」→「認証情報」
4. 「OAuth 2.0 クライアントID」を作成
5. アクセストークンとリフレッシュトークンを取得

### 2. GitHub Secretsの設定

リポジトリの Settings → Secrets and variables → Actions で以下を追加:

- `GOOGLE_ACCESS_TOKEN`: Googleアクセストークン
- `GOOGLE_REFRESH_TOKEN`: Googleリフレッシュトークン

## ローカルでの実行

### Go版を使用する場合 (推奨)

```bash
# 依存関係のインストール
make deps

# ビルド
make build

# 実行 (環境変数を設定して)
export GOOGLE_ACCESS_TOKEN="your_token"
export GOOGLE_REFRESH_TOKEN="your_refresh_token"
make run

# または直接実行
./bin/notebooklm-audio-generator
```

### Python版を使用する場合

```bash
# 依存関係のインストール
pip install -r scripts/requirements.txt

# 実行
python scripts/add_to_notebooklm.py
```

## 開発

### Go版の開発

```bash
# コードフォーマット
make fmt

# リンター実行
make lint

# テスト実行
make test

# ビルド成果物のクリーンアップ
make clean
```

## GitHub Actions

プロジェクトには2つのワークフローがあります:

- `.github/workflows/notebooklm-go.yml` - Go版 (デフォルト)
- `.github/workflows/notebooklm.yml` - Python版

Go版のワークフローがデフォルトで有効になっています。Python版を使用したい場合は、Go版のワークフローを無効化してください。

## 注意事項

- 初回はSeleniumのセレクタ調整が必要な場合があります
- NotebookLMのUI変更に応じてスクリプト修正が必要な場合があります
- Go版はchromedpを使用しているため、ChromeDriverの手動インストールは不要です
