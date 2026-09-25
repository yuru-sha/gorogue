# Core コンポーネント

このガイドは現行 Go パッケージを示します。要件は [`../../SPEC.md`](../../SPEC.md)、コードが実際の API と挙動の根拠です。

- `internal/core/engine.go`: `gruid.Model` 境界、状態・ゲーム・UI の接続。
- `internal/core/state/`: 画面状態と現在画面への入力・描画委譲。
- `internal/core/command/`: Rogue キーや名前付き CLI 入力の解析、および GUI/CLI 共通のゲーム操作実行。
- `internal/core/cli/`: 名前付き CLI コマンドのフロントエンド。
- `internal/core/wizard/`: 開発用ウィザード操作。

ゲームルールは `internal/game/` と `internal/core/command/Executor` に置きます。CLI や画面に同じルールを複製せず、共有コマンド経路を使用します。自動経路探索、MP、魔法詠唱、罠解除を前提にした設計はありません。
