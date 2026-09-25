# GoRogue 品質保証

ゲーム要件は [`../SPEC.md`](../SPEC.md)、実際の挙動は Go コードとテストを正とします。

## ローカル検証

```sh
go test ./...
go vet ./...
make ci-checks
git diff --check
```

`make ci-checks` は `make lint` と `make test` を実行します。`make lint` は `go fmt ./...` で整形した後、`go vet`、golangci-lint、staticcheck を実行します。開発ツールのセットアップは [`development.md`](development.md) を参照してください。

## 変更の検証

- 変更したルールのテストと該当パッケージのテストを先に実行する。
- 完了前に `go test ./...`、`go vet ./...` を実行する。
- 開発ツールが利用可能な場合は `make ci-checks` を実行する。
- シード再現性、セーブ/ロード、勝利/死亡、GUI/CLI 共通コマンド境界を変更に応じて確認する。
- GUI、SDL2、OS 固有、CI 外の挙動を実行できない場合は未検証として報告する。

Makefile に定義されていない QA ターゲット、pre-commit 手順、存在しない CLI シナリオを前提にしません。
