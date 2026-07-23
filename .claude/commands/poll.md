---
description: Notion「記事翻訳キュー」の未処理URLを翻訳して本文に保存する
---

Notion の「記事翻訳キュー」DB にある **未処理の記事 URL** を順に処理する。
1 回の実行で未処理分をすべて処理する。以下を厳密に実行すること。

## 対象（このファイルが運用値の単一情報源）
- data source: `collection://b87c8503-94f1-4e6e-bcf2-1cda324267b6`
- Status 値: 未処理 / 処理中 / 完了 / エラー
- プロパティ更新キー: `Title` / `userDefined:URL` / `Status` / `date:処理日時:start` / `date:処理日時:is_datetime` / `エラー`
- 記事抽出: リポジトリ直下の `.venv/bin/python src/extract.py "<URL>"`（要 `make setup`）

## 手順

### 1. 未処理行の取得
`query-data-sources`（SQL モード）で未処理を取得する:

```sql
SELECT url, "Title", "userDefined:URL", "Status"
FROM "collection://b87c8503-94f1-4e6e-bcf2-1cda324267b6"
WHERE "Status" = '未処理' OR "Status" IS NULL OR "Status" = ''
```

- 返る `url` 列は **ページ ID（page_id）**。以降の更新で使う。
- `userDefined:URL` 列が **翻訳対象の記事 URL**。
- 0 件なら「未処理の記事はありません」と報告して終了。

### 2. 各行をループ処理（1 件ずつ、失敗しても次へ）

**(a) 処理中にする** — `notion-update-page`:
- `page_id`: 行の `url`
- `command`: `update_properties`
- `properties`: `{"Status": "処理中"}`

**(b) 本文抽出** — Bash:
```bash
.venv/bin/python src/extract.py "<記事URL>"
```
- 成功: stdout の JSON から `title` と `text` を得る。
- 失敗（非ゼロ終了）: stderr のメッセージを控え、**(e) エラー処理**へ。

**(c) 翻訳＋要約（あなた＝Claude が自分で行う。外部 CLI は使わない）**
抽出した `text` を **自然で読みやすい日本語**に翻訳する。
- 直訳でなく自然な日本語にする。段落・見出し・リスト構造は保持する。
- 原文の情報を足さない・省かない。
- 本文が長い場合も省略せず全文訳す。
- 訳文をもとに **3〜5 行の日本語要約**を作る。

次の Markdown を本文として組み立てる（`<要約>` `<全文>` を差し替え）:
```
## 要約

<要約>

## 全文（日本語訳）

<全文>
```

**(d) Notion へ書き込み（2 コール）**
1. 本文を挿入 — `notion-update-page`:
   - `page_id`: 行の `url`
   - `command`: `insert_content`
   - `position`: `{"type": "end"}`
   - `content`: 上で組み立てた Markdown
2. プロパティ更新 — `notion-update-page`:
   - `page_id`: 行の `url`
   - `command`: `update_properties`
   - `properties`:
     - `Title`: 抽出した `title`（空なら記事 URL を入れる）
     - `Status`: `完了`
     - `date:処理日時:start`: 現在時刻 ISO8601（`date -u +%Y-%m-%dT%H:%M:%SZ` で取得）
     - `date:処理日時:is_datetime`: `1`
     - `エラー`: `null`（既存のエラーがあれば消す）

**(e) エラー処理**（抽出・翻訳・書き込みのいずれかで失敗した行）
`notion-update-page`:
- `page_id`: 行の `url`
- `command`: `update_properties`
- `properties`: `{"Status": "エラー", "エラー": "<失敗理由を簡潔に>"}`
そのまま次の行へ進む。

### 3. 最後にまとめを報告
処理した件数 / 完了 / エラー を日本語で簡潔に報告する。

## 注意
- 既に `完了` / `処理中` の行は取得対象外（冪等）。
- `処理中` のまま残った行があれば、前回の異常終了の可能性。手動で `未処理` に戻せば再処理される。
- Notion への書き込みはすべて MCP 経由（インテグレーショントークンは不要）。
