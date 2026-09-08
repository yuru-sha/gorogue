# AGENTS.md

GoRogueで作業するエージェント向けの運用規約。ゲーム要件は [SPEC.md](SPEC.md)、実装の詳細はコードと関連ドキュメントを参照する。

## Sources of truth

作業開始時に、対象に応じて次を読む。

- 要件、範囲、除外、受け入れ条件: [SPEC.md](SPEC.md)
- 現在のアーキテクチャ: [docs/architecture.md](docs/architecture.md)
- 開発コマンド: [docs/development.md](docs/development.md)、[Makefile](Makefile)、[go.mod](go.mod)
- 実際の挙動: `cmd/`、`internal/`、`*_test.go`

`SPEC.md`をゲーム要件の正とする。`docs/`の記述が実装やテストと異なる場合、現在の挙動はコードとテストを基準にする。要件を変える必要がある場合は、実装より先に`SPEC.md`と受け入れ条件を更新する。

## Workflow

1. `rtk git status --short --branch`と差分を確認し、既存のユーザー変更を保持する。
2. 関係する`SPEC.md`の節、呼び出し元、関連テストを読む。
3. 既存の型・ヘルパー・依存関係を再利用し、仕様に必要な最小範囲だけ変更する。
4. 変更後に対象テストと`rtk go test ./...`を実行する。開発ツールが揃っている場合は`rtk make ci-checks`も実行する。
5. 最終差分、ワークツリー、検証結果、残るリスクを確認してから報告する。

## Safety boundaries

- 変更対象はこのリポジトリに限る。外部サービス、GitHub、認証情報、ブラウザ、実機は明示的に依頼された場合だけ操作する。
- 明示的に依頼されるまで、コミット、プッシュ、ブランチ作成、PR、マージ、リリースを行わない。
- `rm -rf`、強制プッシュ、履歴書き換え、データベースリセットなど、復旧しにくい操作は行わない。
- 秘密情報、トークン、認証情報、個人情報を出力・コミット・送信しない。`.env`は値を表示せずに使用できる。
- 生成物、バイナリ、カバレッジ、ロックファイル、未追跡ファイルを変更する前に、対象範囲を確認する。
- シェルコマンドは`rtk`経由で実行し、ファイル編集は`apply_patch`を使う。

## Project constraints

- `go.mod`のGoバージョンと既存のGo Modules依存関係を基準にする。新しい依存関係やフレームワークは、必要性と代替案を説明してから追加する。
- `SPEC.md`のGoRogue要件と除外事項に従い、隣接する未依頼機能を追加しない。
- ゲームルールは`internal/core/`と`internal/game/`に集約し、`cmd/`や`internal/ui/`にCLI・GUI固有の重複ルールを追加しない。コマンド処理は既存の`internal/core/command/`と共有経路を優先する。
- ダンジョン生成、戦闘、配置の乱数はゲームが管理する乱数源を使い、同じシードと入力列の再現性を維持する。
- セーブ形式を変更する場合は`internal/game/save/`の`SaveVersion`、互換性チェック、変換処理、テストを同時に更新する。
- Goコードは`gofmt`の形式と既存パッケージ構成に従う。公開APIは必要な場合を除いて壊さない。

## Definition of done

次を満たすまで完了と報告しない。

- 差分が`SPEC.md`の要件・除外事項を満たす。
- 対象テストと可能なら`rtk go test ./...`、`rtk go vet ./...`、`rtk make ci-checks`が成功する。
- 影響を受けるシード、セーブ/ロード、勝敗、CLI/GUI共有ルールの境界を確認する。
- コード、コマンド、設定、公開APIを変えた場合は関連ドキュメントを更新する。
- `rtk git diff --check`が成功し、秘密情報や意図しない変更・未追跡ファイルがない。
- GUI、SDL2、OS差異、外部CIを検証していない場合は、その範囲を明記する。

## Verification entry point

```bash
rtk go test ./...
rtk go vet ./...
rtk make ci-checks  # 開発ツールとセットアップ済みの場合
```

`make ci-checks`は`make lint`と`make test`を実行する。`make setup`や`make setup-dev`は依存関係・開発ツール・SDL2環境を変更するため、必要な場合だけ実行する。

## Git

ステージ済み、未ステージ、未追跡のユーザー変更を整理・削除・stashしない。レビュー指摘や検証失敗への恒久対応は、テスト、検証コマンド、ドキュメントのうち最小の適切な場所に追加する。
