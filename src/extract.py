#!/usr/bin/env python3
"""記事 URL から本文とタイトルを抽出する。

使い方:
    python src/extract.py <URL>

成功時: {"title": ..., "text": ..., "url": ...} を JSON で stdout に出力し exit 0。
失敗時: エラー理由を stderr に出力し、非ゼロ終了（呼び出し側が Status=エラー に反映する）。
"""
import json
import sys

import trafilatura


def extract(url: str) -> dict:
    downloaded = trafilatura.fetch_url(url)
    if not downloaded:
        raise RuntimeError(f"URL を取得できませんでした（到達不可 / タイムアウト等）: {url}")

    data = trafilatura.bare_extraction(
        downloaded,
        with_metadata=True,
        include_comments=False,
        include_tables=True,
        favor_recall=True,
        as_dict=True,
    )
    if not data:
        raise RuntimeError(
            "本文を抽出できませんでした（JS 必須ページ / ペイウォール / 本文が短すぎる可能性）"
        )

    text = (data.get("text") or "").strip()
    if not text:
        raise RuntimeError("本文が空でした")

    title = (data.get("title") or "").strip()
    return {"title": title, "text": text, "url": url}


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: python src/extract.py <URL>", file=sys.stderr)
        return 64
    try:
        out = extract(sys.argv[1])
    except Exception as e:  # noqa: BLE001 - 呼び出し側にメッセージを渡すため広く捕捉
        print(str(e), file=sys.stderr)
        return 1
    json.dump(out, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
