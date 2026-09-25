# GoRogue 技術仕様

本書は実装の位置を案内するための索引です。ゲームの規範要件と受け入れ条件は [`SPEC.md`](../SPEC.md)、実際の API と挙動は Go コードおよびテストです。ここに旧言語の擬似コードや過去のアーキテクチャ案を要件として残しません。

## 実行経路

- GUI: `cmd/gorogue/` → `internal/core/Engine` → `internal/core/state/` → `internal/ui/screen/`。
- CLI: `cmd/gorogue-cli/` → `internal/core/cli/`。
- 共通の操作: `internal/core/command/Parser` が入力を `Command` に変換し、`Executor` がゲーム規則を実行。
- 表示: `internal/ui/screen/` が状態から論理表示セルを構築し、`TextRenderer` が文字、色、gruid `Grid`/`Cell` を作る。

## ゲームデータ

- アクター: `internal/game/actor/`。
- ダンジョン、階層、地形、生成、視界、罠: `internal/game/dungeon/`。
- アイテム: `internal/game/item/`。
- 魔法効果と識別: `internal/game/magic/`, `internal/game/identification/`。
- 保存・読込: `internal/game/save/`。`SaveVersion` がバージョンの唯一の定義。

ゲーム規則の乱数はゲーム管理の乱数源を経由します。GUI と CLI は別のルール実装を持ちません。

## 検証

```sh
go test ./...
go vet ./...
make ci-checks
```

個別受け入れ条件と出典は [`SPEC.md`](../SPEC.md) を参照してください。
