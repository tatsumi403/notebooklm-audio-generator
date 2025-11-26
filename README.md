# NotebookLM Audio Generator

URLを追加するだけでNotebookLMに自動的にソースとして登録し、音声解説を生成します。

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

## 注意事項

- 初回はSeleniumのセレクタ調整が必要な場合があります
- NotebookLMのUI変更に応じてスクリプト修正が必要な場合があります
