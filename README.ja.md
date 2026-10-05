# GoRogue

[English](README.md) | [日本語](README.ja.md)

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/yuru-sha/gorogue)

GoRogue は Go と gruid で実装した、Rogue 5.4.4 を基準とする文字表示のローグライクゲームです。要件と受け入れ条件は [`docs/SPEC.md`](docs/SPEC.md) を参照してください。

## 現行機能

- 26 階のシード付きダンジョン、部屋・通路・暗室・宝物室・迷路階・隠し通路/扉。
- Rogue に沿った移動、戦闘、成長、空腹、モンスター、装備、アイテム、識別、罠。
- GUI と CLI で共有するコマンド実行経路。
- アミュレットを地上へ持ち帰る勝利条件とパーマデス。
- JSON セーブ形式 `1.5.0`。

## 操作と画面表示

| キー | 操作 |
| --- | --- |
| `h j k l y u b n` | 8 方向へ移動 |
| `H J K L Y U B N` | 8 方向へ走る |
| `.` | 休む |
| `f` | 方向を指定して戦う |
| `,` | アイテムを拾う |
| `d` | アイテムを落とす |
| `e` | 食べる |
| `w` / `W` | 武器を装備 / 防具を着る |
| `T` | 防具を脱ぐ |
| `P` / `R` | 指輪を装備 / 指輪を外す |
| `t` | アイテムを投げる |
| `q` / `r` / `z` | 薬を飲む / 巻物を読む / 杖を振る |
| `s` / `^` | 周囲を探す / 指定方向を調べる |
| `i` / `@` | 所持品 / キャラクター情報 |
| `D` / `c` | 発見済みアイテム / 未識別アイテムに名前を付ける |
| `<` / `>` | 上る / 下りる |
| `a` | 直前のコマンドを繰り返す |
| `?` / `/` | コマンドの説明 / 記号の説明 |
| `Q` / `Escape` | 終了 / キャンセル |
| `^L` | セーブデータをロードする |

ゲームはダンジョン内から始まります。画面上部に最新のメッセージや入力プロンプト、中央にマップ、最下行に階層、金貨、HP、力、鎧、経験値を表示します。空腹時は最下行に `Hungry`、`Weak`、`Faint` を表示します。`?` の後にキーを押すとコマンドの説明を表示し、`*` で一覧を表示します。死亡と勝利は段階表示の後にゲームを終了します。

GoRogue 固有の補助操作として `^W` でウィザードモード、`^L` でセーブデータをロードします。`:` で CLI デバッグ入力に切り替えます。CLI からもセーブとロードを実行でき、SIGHUP を受けると復旧用セーブを保存して終了します。

## ユーザーオプション

環境変数または `.env` で `ROGUEOPTS` を設定します。値はカンマ区切りで、例は `name=Ada,noflush,inven=clear,file=~/rogue.sav` です。GUI では `O` で設定を表示・編集できます。各 option と GUI での対応は [`docs/features.md`](docs/features.md#rogue-544-user-options) を参照してください。

## ビルドと実行

```sh
# 開発ツールと SDL2 の確認・セットアップ
make setup-dev

# SDL2 GUI
make build
make run

# 固定シードの CLI
go run ./cmd/gorogue-cli --seed 12345

# テスト・静的解析
go test ./...
go vet ./...
```

Go バージョンとモジュール依存関係は [`go.mod`](go.mod) を正とします。開発コマンドの詳細は [`docs/development.md`](docs/development.md)、実装構成は [`docs/architecture.md`](docs/architecture.md) を参照してください。


## GitHub運用

Issue Form と既定の Pull Request テンプレートは `yuru-sha/.github` の共通設定を利用します。Release Notes のカテゴリは `.github/release.yml` で管理し、`orca:*` を含む共通ラベルは `yuru-sha/project-template` から同期します。
