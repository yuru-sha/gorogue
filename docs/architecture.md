# GoRogue アーキテクチャ

## 規範と境界

ゲーム要件は [`SPEC.md`](../SPEC.md)、コードとテストは現行挙動の根拠です。Rogue 5.4.4 のルールはゲーム層と共有コマンド経路に置き、画面や CLI に複製しません。

```text
GUI / CLI 入力
      ↓
internal/core/command (Parser → Executor)
      ↓
internal/game (actor, dungeon, item, magic, save)
      ↓
ゲーム状態
      ↓
internal/ui/screen の論理表示セル変換
      ↓
TextRenderer (記号・色・gruid Grid/Cell)
      ↓
GUI ドライバーまたは端末表示
```

## パッケージ責務

- `internal/core/command/`: Rogue キーと CLI 命令を共有 `Command` に解析し、共通 `Executor` でルールを実行します。
- `internal/core/state/`: 画面状態の切り替えと、選択中画面への入力・描画の委譲を行います。
- `internal/core/`: `Engine` がゲーム状態、セーブ連携、gruid モデル境界を接続します。画面グリッドは `internal/ui/screen.TextRenderer` が所有します。
- `internal/game/actor/`: プレイヤー、モンスター、戦闘・状態。
- `internal/game/dungeon/`: 26 階、タイル、部屋、罠、生成、可視性、階層間移動。
- `internal/game/item/`, `magic/`, `identification/`: アイテム定義、効果、外観、識別。
- `internal/game/save/`: JSON 保存、読込、検証、変換。現行 `SaveVersion` は `1.5.0`。
- `internal/ui/screen/`: 入力画面、論理表示セルの生成、文字・色とグリッドへの描画。
- `cmd/gorogue/`, `cmd/gorogue-cli/`: GUI と CLI の起動。

## ゲーム状態と表示

ルール型は gruid の `Grid`、`Cell`、色、文字グリフを保持しません。タイルは論理地形、可視性、探索状態を保持します。スクリーン層はこれらを画面位置・地形・可視/探索状態・表示エンティティ・優先度を持つ論理表示セルに変換します。表示セルから terminal glyph/color への変換は文字レンダラーの責務です。プレイヤー、モンスター、罠、アイテムの重なり順と隠れた対象の非表示は変換時に決定します。

この境界は文字表示を対象とし、画像タイルやアニメーションは導入しません。

## 再現性と保存

ダンジョン生成、配置、戦闘、アイテム効果はゲーム管理の乱数源を使います。同一バージョン、同一シード、同一入力列で結果を再現できることが契約です。階層、地形、アイテム、モンスター、罠、プレイヤー状態、識別状態、乱数カーソルなど、進行復元に必要な状態を JSON に保存します。旧保存形式の自動移行は保証しません。

保存形式を変更するときは `SaveVersion`、互換性判定、変換、保存/読込テストを同時に更新します。
