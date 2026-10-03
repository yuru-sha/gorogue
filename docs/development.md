# GoRogue 開発者ガイド

## 必要環境

- Go バージョンと依存関係は [`go.mod`](../go.mod) を正とする。
- `make`。
- SDL2 開発ライブラリと `pkg-config`。GUI のビルド/実行に必要。
- `make ci-checks` では Makefile に定義された golangci-lint と staticcheck のバージョンを使う。

macOS では Homebrew で SDL2 を用意できます。

```sh
brew install pkg-config sdl2
```

## 実行

```sh
# SDL2 GUI
make setup-dev
make run

# CLI
go run ./cmd/gorogue-cli --seed 12345

# CLI のオプション
go run ./cmd/gorogue-cli --help
```

`make setup-dev` は開発用ツールをインストールし、SDL2 とツールのバージョンを検査します。`make setup` は Go module のダウンロードと tidy を実行します。依存関係の変更が必要でない限り、セットアップターゲットをむやみに再実行しないでください。

## 検証

```sh
go test ./...
go vet ./...
make ci-checks
git diff --check
```

`make ci-checks` は `make lint` と `make test` を実行します。`make lint` は `go fmt ./...` により Go ファイルを整形してから `go vet`、golangci-lint、staticcheck を実行します。CI は [`quality.yml`](../.github/workflows/quality.yml) に定義されています。SDL2 GUI の起動はヘッドレス CI では検証しません。

## シードと保存

CLI は `--seed` で乱数シードを指定できます。ゲーム API にはシード指定コンストラクターがあります。同一バージョン、同一シード、同一入力列の再現性を維持してください。生成、配置、戦闘、アイテム効果の乱数にはゲーム管理の乱数源を使用します。

セーブ形式の現行バージョンは `internal/game/save/` の `SaveVersion` が正です。形式を変える場合は、互換性判定、変換、保存/読込テスト、関連ドキュメントを一緒に更新してください。旧形式の自動移行は保証しません。

## 保存・復元ライフサイクル

Rogue 5.4.4 の基準実装は [`save.c` (`cf9bd26`)](https://github.com/Davidslv/rogue/blob/cf9bd26/save.c) です。GoRogue は保存 JSON と `SaveVersion` を維持し、手動保存の永続化成功後にプレイを終了します。保存に失敗した場合はプレイを継続します。復元では、互換性検証とゲーム状態への変換を終えてから保存を消費し、その成功後に新しい状態を公開します。検証、変換、保存の消費のいずれかに失敗した場合は、現在の状態と保存ファイルを維持します。復元に成功すると保存を消費するため、同じ保存を再度復元できません。

GUI と CLI は SIGHUP を受けると、進行中のゲームをターンを進めずに復旧用セーブへ保存して終了します。保存に失敗した場合も終了します。ゲーム開始前の GUI メニューでは既存のセーブを上書きせずに終了します。Rogue の `auto_save` と同様に終了を優先しますが、復元時の保存消費は同実装の早期 unlink を踏襲せず、すべての検証と変換が成功した後に行います。

## コード配置

- ゲームルール: `internal/game/` と `internal/core/command/`。
- 共有入力と実行: `internal/core/command/`。
- GUI/CLI 固有処理: `cmd/` と `internal/ui/`。
- テスト: 対象パッケージの `*_test.go`。

GUI と CLI のルールを重複実装しないでください。ゲーム状態に表示グリフ、色、gruid `Grid`/`Cell` を追加せず、表示層の論理セル変換と文字レンダラーを利用してください。
