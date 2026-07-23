.PHONY: setup extract clean help

VENV := .venv
PY := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

# venv を作り依存をインストール
setup:
	@test -d $(VENV) || python3 -m venv $(VENV)
	@$(PIP) install --upgrade pip >/dev/null
	@$(PIP) install -r requirements.txt
	@echo "setup done. -> source $(VENV)/bin/activate"

# 記事抽出の単体テスト:  make extract URL=https://example.com/article
extract:
	@test -n "$(URL)" || (echo "usage: make extract URL=<記事URL>" && exit 64)
	@$(PY) src/extract.py "$(URL)"

clean:
	@rm -rf $(VENV) src/__pycache__

help:
	@echo "make setup            - .venv 作成 + 依存インストール"
	@echo "make extract URL=...  - 記事抽出の単体テスト"
	@echo "make clean            - .venv とキャッシュ削除"
	@echo ""
	@echo "翻訳して Notion に保存するには Claude Code で /poll を実行してください。"
