# NotebookLM Audio Generator - 実装手順書

## プロジェクト概要

URLをリポジトリに追加すると、GitHub ActionsがSeleniumでNotebookLMにアクセスし、ソースとして追加して音声解説を自動生成するシステム。

## リポジトリ構成

```
notebooklm-audio-generator/
├── .github/
│   └── workflows/
│       └── notebooklm.yml          # GitHub Actionsワークフロー
├── scripts/
│   ├── add_to_notebooklm.py       # メインスクリプト
│   └── requirements.txt            # Python依存パッケージ
├── urls.txt                        # URL管理ファイル
├── .processed_urls.txt             # 処理済みURL記録(自動生成)
└── README.md                       # 使い方説明
```

## 実装ステップ

### 1. `urls.txt` の作成

```txt
# NotebookLMに追加したいURLを1行ずつ記載
# 例:
https://example.com/article1
https://example.com/article2
```

### 2. `scripts/requirements.txt` の作成

```txt
selenium==4.15.0
webdriver-manager==4.0.1
google-auth==2.23.4
google-auth-oauthlib==1.1.0
google-auth-httplib2==0.1.1
```

### 3. `scripts/add_to_notebooklm.py` の作成

```python
import os
import sys
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager
import time
import pickle

def setup_driver():
    """ヘッドレスChromeの設定"""
    chrome_options = Options()
    chrome_options.add_argument('--headless')
    chrome_options.add_argument('--no-sandbox')
    chrome_options.add_argument('--disable-dev-shm-usage')
    chrome_options.add_argument('--disable-gpu')
    chrome_options.add_argument('--window-size=1920,1080')
    
    service = Service(ChromeDriverManager().install())
    driver = webdriver.Chrome(service=service, options=chrome_options)
    return driver

def load_cookies(driver):
    """保存されたCookieを読み込む"""
    cookie_file = os.getenv('GOOGLE_COOKIES_FILE', 'cookies.pkl')
    if os.path.exists(cookie_file):
        with open(cookie_file, 'rb') as f:
            cookies = pickle.load(f)
            for cookie in cookies:
                driver.add_cookie(cookie)
        return True
    return False

def login_with_oauth(driver):
    """OAuth認証でログイン"""
    # この部分は手動で初回認証を行い、Cookieを保存する必要がある
    # GitHub Secretsに保存したトークン情報を使用
    access_token = os.getenv('GOOGLE_ACCESS_TOKEN')
    refresh_token = os.getenv('GOOGLE_REFRESH_TOKEN')
    
    if not access_token:
        print("Error: GOOGLE_ACCESS_TOKEN not found in environment")
        sys.exit(1)
    
    # NotebookLMにアクセス
    driver.get('https://notebooklm.google.com')
    time.sleep(3)
    
    # アクセストークンを使用したログイン処理
    # (実際の実装はNotebookLMの認証フローに依存)
    driver.execute_script(f"""
        localStorage.setItem('access_token', '{access_token}');
        localStorage.setItem('refresh_token', '{refresh_token}');
    """)
    
    driver.refresh()
    time.sleep(3)

def add_url_to_notebooklm(driver, url):
    """URLをNotebookLMのソースとして追加"""
    try:
        # NotebookLMのプロジェクトページへ移動
        # (実際のセレクタはNotebookLMのUIに合わせて調整が必要)
        
        # 「ソースを追加」ボタンをクリック
        add_source_btn = WebDriverWait(driver, 10).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(., 'Add source')]"))
        )
        add_source_btn.click()
        time.sleep(2)
        
        # URL入力フィールドを探して入力
        url_input = WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, "input[type='url']"))
        )
        url_input.send_keys(url)
        time.sleep(1)
        
        # 追加ボタンをクリック
        submit_btn = driver.find_element(By.XPATH, "//button[contains(., 'Add')]")
        submit_btn.click()
        time.sleep(5)
        
        print(f"Successfully added: {url}")
        return True
    except Exception as e:
        print(f"Error adding {url}: {str(e)}")
        return False

def generate_audio_guide(driver):
    """音声ガイドを生成"""
    try:
        # Studioボタンをクリック
        studio_btn = WebDriverWait(driver, 10).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(., 'Studio')]"))
        )
        studio_btn.click()
        time.sleep(2)
        
        # 音声生成ボタンをクリック
        generate_btn = WebDriverWait(driver, 10).until(
            EC.element_to_be_clickable((By.XPATH, "//button[contains(., 'Generate')]"))
        )
        generate_btn.click()
        
        print("Audio generation started")
        time.sleep(10)  # 生成完了を待つ
        return True
    except Exception as e:
        print(f"Error generating audio: {str(e)}")
        return False

def get_new_urls():
    """新規追加されたURLを取得"""
    with open('urls.txt', 'r') as f:
        all_urls = [line.strip() for line in f if line.strip() and not line.startswith('#')]
    
    processed_file = '.processed_urls.txt'
    if os.path.exists(processed_file):
        with open(processed_file, 'r') as f:
            processed_urls = set(line.strip() for line in f)
    else:
        processed_urls = set()
    
    new_urls = [url for url in all_urls if url not in processed_urls]
    return new_urls, processed_file

def mark_as_processed(url, processed_file):
    """URLを処理済みとしてマーク"""
    with open(processed_file, 'a') as f:
        f.write(f"{url}\n")

def main():
    print("Starting NotebookLM automation...")
    
    # 新規URLを取得
    new_urls, processed_file = get_new_urls()
    
    if not new_urls:
        print("No new URLs to process")
        return
    
    print(f"Found {len(new_urls)} new URLs to process")
    
    # ドライバーを起動
    driver = setup_driver()
    
    try:
        # ログイン
        login_with_oauth(driver)
        
        # 各URLを処理
        for url in new_urls:
            print(f"Processing: {url}")
            if add_url_to_notebooklm(driver, url):
                mark_as_processed(url, processed_file)
                time.sleep(2)
        
        # 音声ガイドを生成
        generate_audio_guide(driver)
        
        print("All URLs processed successfully")
        
    except Exception as e:
        print(f"Error in main process: {str(e)}")
        sys.exit(1)
    finally:
        driver.quit()

if __name__ == "__main__":
    main()
```

### 4. `.github/workflows/notebooklm.yml` の作成

```yaml
name: Add URLs to NotebookLM

on:
  push:
    branches:
      - main
    paths:
      - 'urls.txt'

jobs:
  process-urls:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v4
      with:
        fetch-depth: 2
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.11'
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r scripts/requirements.txt
    
    - name: Install Chrome
      run: |
        sudo apt-get update
        sudo apt-get install -y google-chrome-stable
    
    - name: Run NotebookLM automation
      env:
        GOOGLE_ACCESS_TOKEN: ${{ secrets.GOOGLE_ACCESS_TOKEN }}
        GOOGLE_REFRESH_TOKEN: ${{ secrets.GOOGLE_REFRESH_TOKEN }}
      run: |
        python scripts/add_to_notebooklm.py
    
    - name: Commit processed URLs
      run: |
        git config --local user.email "github-actions[bot]@users.noreply.github.com"
        git config --local user.name "github-actions[bot]"
        git add .processed_urls.txt
        git diff --quiet && git diff --staged --quiet || git commit -m "Update processed URLs [skip ci]"
        git push
```

### 5. `README.md` の作成

```markdown
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
```

## 次のステップ

1. **ローカルで動作確認**: まずローカル環境でSeleniumスクリプトを実行し、NotebookLMの正確なセレクタを特定
1. **OAuth認証設定**: Google Cloud Consoleでプロジェクト作成とOAuth設定
1. **GitHub Secrets設定**: 認証情報を安全に保存
1. **動作テスト**: urls.txtにテストURLを追加してpush

## 重要な調整ポイント

`add_to_notebooklm.py` 内の以下の部分は、実際のNotebookLMのUIに合わせて調整が必要:

- セレクタ (XPath, CSS Selector)
- ボタンやフィールドのテキスト
- 待機時間 (time.sleep)

実際のNotebookLM画面を見ながら、Chrome DevToolsで要素を特定してください。
