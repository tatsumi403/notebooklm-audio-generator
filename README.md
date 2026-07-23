# 記事翻訳 → Notion（英語記事を自然な日本語にして貯める）

英語記事の URL を Notion に貼るだけで、自然な日本語訳（冒頭に要約付き）を
その Notion ページに保存するツール。翻訳の実行は Claude Code の **`/poll`** コマンドで都度行う。

> 旧版は Google NotebookLM をブラウザ自動操作して音声生成する設計だったが、
> OAuth が使えず動かなかったため、この構成に作り直した（履歴は git に残っている）。

## 使い方（日常運用）

1. Notion の DB **「記事翻訳キュー」** に行を追加し、`URL` 列に英語記事の URL を貼る
   （PC でもスマホでも可。何件でも溜めてよい）。
   - DB: https://app.notion.com/p/8ee50b66c1c34c5494050d5b5a6c5174
2. 翻訳したくなったら、このリポジトリで Claude Code を開き **`/poll`** と入力する。
3. 未処理の行がまとめて処理され、各ページ本文に「要約 → 全文（日本語訳）」が入り、
   `Status` が `完了` になる。さらに `tatsumi403/mylife` に記録用の GitHub Issue が作られ、
   本文に生成した Notion ページ URL が貼られる（Project「my life ロードマップ」/ Priority「1週間以内」）。

トリガーはこの `/poll` だけ。ポーリングや常駐はしない（打った時だけ動く）。

## 仕組み

```
[Notion DB に URL を貼る (Status=未処理)]
            │  ← 好きなタイミングで /poll
            ▼
[Claude Code /poll]
  1. Notion MCP で未処理行を取得
  2. Status=処理中 に更新
  3. src/extract.py で記事本文を抽出（trafilatura）
  4. Claude 自身が自然な日本語に翻訳＋要約
  5. Notion MCP でページ本文に要約＋全文を書き込み
  6. Status=完了・処理日時 を記録（失敗時は Status=エラー）
  7. mylife に記録用 GitHub Issue を作成（Notion URL を貼付／Project・Priority を設定）
```

- **翻訳**は Claude Code 自身が行う（外部 API キー不要）。
- **Notion 連携**は Claude Code の MCP コネクタ経由（インテグレーショントークン不要）。
- **記事抽出**だけ Python（`trafilatura`）に切り出している。
- **記録用 Issue** を `tatsumi403/mylife` に作成し、生成した Notion ページ URL を本文に貼る
  （Project「my life ロードマップ」／ Priority「1週間以内」。アサインと Status は Project 側の
  自動化で設定される）。Issue 作成には `gh` の `project` スコープが必要（`gh auth refresh -s project`）。

## セットアップ

```bash
make setup          # .venv 作成 + trafilatura インストール
```

Notion 側は既に DB 作成済み。MCP コネクタが接続済みであればトークン設定は不要。

## 単体テスト

```bash
make extract URL=https://example.com/some-article   # 記事抽出だけ確認（JSON 出力）
```

翻訳〜Notion 保存まで通すには Claude Code で `/poll` を実行する。

## ファイル構成

| パス | 役割 |
|---|---|
| `.claude/commands/poll.md` | `/poll` の手順書（実行の本体。運用値の単一情報源） |
| `src/extract.py` | 記事 URL → 本文/タイトル抽出（trafilatura） |
| `requirements.txt` | Python 依存（trafilatura） |
| `Makefile` | `setup` / `extract` |

## 今後（フェーズ2・未実装）

要約＋全文を素材に「2 人の対話台本」を生成し、TTS で音声（ラジオ風）にして
Notion に添付する拡張を想定している。
