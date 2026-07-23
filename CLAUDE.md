# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Important Instructions

- **Responses must be in Japanese**: All communication with users must be in Japanese.

## Project Overview

英語記事を自然な日本語に翻訳して Notion に保存するツール。ユーザーは Notion の DB「記事翻訳キュー」に
記事 URL を貼るだけ。翻訳の実行は Claude Code の **`/poll`** コマンドで都度行う（ポーリング・常駐はしない）。

> 旧版は Google NotebookLM をブラウザ自動操作（Go+chromedp / Python+Selenium）して音声生成する設計だったが、
> OAuth が取得できず稼働しなかったため全面的に作り直した。旧実装は git 履歴にのみ残る。

## アーキテクチャ（重要）

処理の**本体は `.claude/commands/poll.md`**（＝Claude への手順書）。`/poll` 実行時に Claude 自身が:
1. Notion MCP で未処理行を取得（`query-data-sources`, SQL モード）
2. `Status=処理中` に更新
3. `src/extract.py` を Bash で実行し記事本文を抽出（trafilatura）
4. **Claude 自身が翻訳＋要約**（外部 API/CLI は使わない。Gemini CLI はヘッドレスで認証不可のため不採用）
5. Notion MCP でページ本文に「要約→全文」を書き込み（`notion-update-page`）
6. `Status=完了`・処理日時 を記録（失敗時は `Status=エラー`＋メッセージ）

設計上の要点:
- **Notion 連携は MCP 経由**。インテグレーショントークンは使わない（発行不可のため）。
  よって Notion I/O は「MCP を持つ Claude（`/poll`）」がやる。素の Python からは呼べない。
- **翻訳は Claude 自身**が担当。`translate.py` や `gemini` 依存は持たない。
- **記事抽出のみ Python**（`src/extract.py`）に切り出し。

## コマンド

```bash
make setup                 # .venv 作成 + trafilatura インストール（初回のみ）
make extract URL=<記事URL> # 記事抽出だけの単体テスト（JSON を stdout）
```

翻訳〜Notion 保存の通し実行は Claude Code で **`/poll`**。`src/extract.py` は `.venv/bin/python` から実行する。

## Notion DB「記事翻訳キュー」

親ページ「黒田_作業ドキュメント」配下に作成済み。**運用値（data source ID・プロパティ名・Status 値）の
単一の情報源は `.claude/commands/poll.md`**（`/poll` は poll.md の手順どおりにしか動かないため）。
ここでは同じ値を再掲せず、扱う上での注意点だけ記す。

**注意:** `url`/`id` という名のプロパティは MCP 上 `userDefined:` 接頭辞が必須（URL 列 → `userDefined:URL`）。
`query-data-sources` の結果 `url` 列は各行の **page_id**（更新時に使用）。`Status` 空 or `未処理` が処理対象、
`完了`/`処理中` はスキップ（冪等）。

## 変更時の注意

- 処理フローや Notion のプロパティ/Status を変えたら、`.claude/commands/poll.md` を更新する（運用値の唯一の置き場）。
- MCP ツール名は Notion コネクタ（`query-data-sources` / `notion-update-page` / `fetch` など）を使用。

## フェーズ2（未実装）

要約＋全文を素材に「2 人の対話台本」を生成 → TTS で音声（ラジオ風）→ Notion 添付、という拡張を想定。
