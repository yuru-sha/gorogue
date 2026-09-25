# GoRogue プロジェクト概要

GoRogue は Go と gruid で実装した、Rogue 5.4.4 を基準とする文字表示のローグライクゲームです。ゲーム要件と受け入れ条件は [`SPEC.md`](../SPEC.md)、実際の挙動はコードとテストを正とします。

## ゲーム

プレイヤーは 26 階のダンジョンを探索し、イェンダーのアミュレットを地上へ持ち帰ります。HP が尽きるとそのゲームは終了します。ダンジョン生成、戦闘、食料、モンスター、アイテム、識別、罠、コマンドは Rogue 5.4.4 の原典に合わせます。

GUI と CLI は同じコマンド実行経路を利用します。CLI とゲーム API はシードを指定でき、同じバージョン、シード、入力列で再現できるゲーム乱数を使います。

## 実行方法

```sh
# SDL2 GUI
make setup-dev
make run

# シードを指定した CLI
go run ./cmd/gorogue-cli --seed 12345

# テストと静的解析
go test ./...
go vet ./...
```

`make setup-dev` は開発用ツールと SDL2 環境を検査・セットアップします。利用可能な検証コマンドは [`docs/development.md`](development.md) を参照してください。

## 実装構成

- `cmd/gorogue/`: SDL2 GUI のエントリーポイント。
- `cmd/gorogue-cli/`: CLI のエントリーポイント。
- `internal/core/command/`: GUI/CLI 共通のコマンド解析と実行。
- `internal/core/`: ゲームエンジン、状態、CLI 統合。
- `internal/game/`: アクター、ダンジョン、アイテム、魔法、セーブ。
- `internal/ui/screen/`: 画面、論理表示セル変換、文字レンダラー。

JSON セーブの現行バージョンは `1.4.0` です。詳細な責務とデータの流れは [`docs/architecture.md`](architecture.md) を参照してください。
