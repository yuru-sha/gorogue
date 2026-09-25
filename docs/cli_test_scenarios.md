# CLI 手動確認シナリオ

CLI の実際のコマンドと引数は `cmd/gorogue-cli/` および `internal/core/cli/` を正とします。自動テストは各 Go パッケージの `*_test.go` にあります。この文書は実行時の手動確認を案内し、存在しない Python CLI やコマンドを記載しません。

## 起動と終了

```sh
go run ./cmd/gorogue-cli --help
go run ./cmd/gorogue-cli --seed 12345
```

対話入力で `help`、`status`、`inventory` を確認し、最後に `quit` を入力します。

## 再現可能な起動

同じ seed を指定して CLI を複数回起動し、初期階層が生成されることを確認します。ゲーム乱数に依存する確認では、同じバージョンと同じ入力列を使います。

## ゲームルール境界

GUI と CLI のコマンドが共有 executor を通ることは `internal/core/command/` のテストと `internal/ui/screen/` のゲームプレイテストで検証します。CLI の確認後は `go test ./internal/core/cli ./internal/core/command ./internal/ui/screen` を実行します。

CI では [`quality.yml`](../.github/workflows/quality.yml) に定義されたヘッドレステストを実行します。GUI ウィンドウや SDL2 の実画面確認は CI の保証範囲ではありません。
