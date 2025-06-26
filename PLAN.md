# Whisperアプリケーション開発計画

## プロジェクトの目標
Go、Wails、Pythonを使用してWhisperアプリケーションを開発する。

## 現在の状況
*   Wailsプロジェクトの基本構造は整っている。
*   `whisper-app-wails/python/main.py`ファイルが不足していたため、プレースホルダーとして作成済み。
*   Wailsプロジェクトのビルドを試みたが、`wails`コマンドが見つからなかった。
*   `wails` CLIのインストールを試みたが、`go`コマンドが見つからなかった。

## 次のステップ（ユーザーの対応が必要）
1.  **Goのインストール:** システムにGoがインストールされていることを確認してください。Goの公式ウェブサイトからダウンロードするか、お使いのOSのパッケージマネージャーを使用してインストールできます。
2.  **Wails CLIのインストール:** Goがインストールされたら、以下のコマンドを実行してWails CLIをインストールしてください。
    ```bash
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    ```

## 次のステップ（GoとWailsがインストールされた後のエージェントの対応）
1.  **Wailsプロジェクトのビルド:** Wails CLIが利用可能になったら、`whisper-app-wails`ディレクトリでプロジェクトをビルドします。
    ```bash
    wails build
    ```
2.  **Python統合（Whisperモデル）:**
    *   `whisper-app-wails/python/main.py`を更新し、Whisperモデルの読み込みと文字起こしロジックを実装します。
    *   必要なPythonライブラリ（`whisper`, `torch`など）のインストール方法を検討します。
3.  **Goバックエンドの実装:**
    *   `whisper-app-wails/app.go`を更新し、音声ファイルの選択、Pythonスクリプトの呼び出し、文字起こし結果の受け渡しを処理するロジックを実装します。
4.  **フロントエンド（Svelte）の実装:**
    *   `whisper-app-wails/frontend/src/App.svelte`を更新し、音声ファイル選択UI、文字起こし中のローディング表示、文字起こし結果の表示を実装します。
5.  **テストと改善:**
    *   基本的なエラー処理を追加し、アプリケーション全体のワークフローをテストします。