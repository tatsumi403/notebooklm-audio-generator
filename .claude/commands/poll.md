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
- GitHub Issue 連携（翻訳完了後に記録用 Issue を作る先）:
  - リポジトリ: `tatsumi403/mylife`（owner: `tatsumi403`）
  - Project: `my life ロードマップ`
  - Priority: 単一選択オプション `1週間以内`（Status・アサインは Project 側の自動化に任せ、**設定しない**）
  - 認証: `gh` に `project` スコープが必要（無ければ `gh auth refresh -s project`）

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

**(準備) GitHub 連携の ID を一度だけ解決**（ループの外で 1 回。得た 4 値を各行の (e) で使い回し、行ごとに再取得しない）:
```bash
# Project 番号(PNUM)と node id(PID)
read -r PNUM PID <<<"$(gh project list --owner tatsumi403 --format json \
  --jq '.projects[] | select(.title=="my life ロードマップ") | "\(.number) \(.id)"')"
# 「1週間以内」オプションを持つ単一選択フィールドの field id と option id
read -r FIELD OPT <<<"$(gh project field-list "$PNUM" --owner tatsumi403 --format json \
  --jq '.fields[] | .id as $fid | .options[]? | select(.name=="1週間以内") | "\($fid) \(.id)"')"
```
`PNUM`/`PID`/`FIELD`/`OPT` のいずれかが空なら Project 名／オプション名が変わった可能性。直すまで各行の (e) はスキップする。

**(a) 処理中にする** — `notion-update-page`:
- `page_id`: 行の `url`
- `command`: `update_properties`
- `properties`: `{"Status": "処理中"}`

**(b) 本文抽出** — Bash:
```bash
.venv/bin/python src/extract.py "<記事URL>"
```
- 成功: stdout の JSON から `title` と `text` を得る。
- 失敗（非ゼロ終了）: stderr のメッセージを控え、**(f) エラー処理**へ。

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

**(e) mylife に記録用 GitHub Issue を作成（(d) の Notion 書き込みが成功した行だけ）**
記録用 Issue を `tatsumi403/mylife` に作り、Project 追加と Priority 設定まで行う。
`NOTION_URL`（翻訳済み記事の Notion ページ URL）は、行の `url` 列（`https://app.notion.com/<id>` 形式）の
ホスト直後に **`/p/` を挿入**した `https://app.notion.com/p/<id>` を使う。`url` 列の値そのままは page_id 用で、
ブラウザでは 404（`This page couldn't be found`）になる（`fetch` が返すページの正規 URL は `/p/` 付き）。
`PNUM`/`PID`/`FIELD`/`OPT` は **(準備)** で解決した値をそのまま使う（Status・アサインは Project 自動化に任せ、指定しない）:
```bash
NOTION_URL="https://app.notion.com/p/<行の url 列の id 部分>"
ART_URL="<記事URL（userDefined:URL 列）>"
ISSUE_URL=$(gh issue create --repo tatsumi403/mylife \
  --title "記事を読む: <記事タイトル: (c) の title。空なら記事URL>" \
  --body "$(printf '翻訳済み記事（Notion）: %s\n\n元記事: %s\n' "$NOTION_URL" "$ART_URL")")
ITEM=$(gh project item-add "$PNUM" --owner tatsumi403 --url "$ISSUE_URL" --format json --jq '.id')
gh project item-edit --id "$ITEM" --project-id "$PID" --field-id "$FIELD" --single-select-option-id "$OPT"
```

Issue 作成〜Priority 設定のどこかで失敗した場合: Notion の翻訳は既に `完了` なので **Status は変えず**、
`エラー` 欄に「Issue作成失敗: <理由>」を記録し、最後のまとめで報告する（後で手動で Issue を作成する）。
本文の二重挿入・再翻訳を招くため、この行を `未処理`/`エラー` に戻さない。

**(f) エラー処理**（抽出・翻訳・Notion 書き込みのいずれかで失敗した行）
`notion-update-page`:
- `page_id`: 行の `url`
- `command`: `update_properties`
- `properties`: `{"Status": "エラー", "エラー": "<失敗理由を簡潔に>"}`
そのまま次の行へ進む。

### 3. 最後にまとめを報告
処理した件数 / 完了 / 作成した Issue 数 / エラー を日本語で簡潔に報告する。
Issue 作成のみ失敗した行（Notion は `完了`）があれば、それも分けて報告する。

## 注意
- 既に `完了` / `処理中` の行は取得対象外（冪等）。
- `処理中` のまま残った行があれば、前回の異常終了の可能性。手動で `未処理` に戻せば再処理される。
- Notion への書き込みはすべて MCP 経由（インテグレーショントークンは不要）。
