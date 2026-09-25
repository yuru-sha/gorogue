# 開発タスクと仕様

このファイルは過去の作業一覧を現在の要件として再利用しません。現行の対象範囲、受け入れ条件、実装順序は [`../SPEC.md`](../SPEC.md)、実際の挙動は Go コードとテストが正です。Rogue 5.4.4 にない機能や旧アーキテクチャ案を完了済みタスクとして記載しません。

## 変更時の確認先

- ダンジョン、プレイヤー、モンスター、アイテム、罠、操作、表示、保存: [`../SPEC.md`](../SPEC.md) の該当節。
- GUI/CLI 共通ルール: `internal/core/command/`。
- 保存形式: `internal/game/save/SaveVersion` と同パッケージの互換性テスト。
- 開発・CI コマンド: [`development.md`](development.md)。
- パッケージと状態/表示の境界: [`architecture.md`](architecture.md)。

## 完了基準

仕様の受け入れ条件に対応する回帰テストを追加し、対象テスト、`go test ./...`、`go vet ./...`、利用可能な場合は `make ci-checks` を実行します。シード、保存/読込、勝利/死亡、GUI/CLI の共有ルール境界を確認し、影響する文書を更新します。
