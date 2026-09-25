# 設定コンポーネント

設定実装は `internal/config/` にあります。設定は環境変数と optional `.env` から読み込み、型付きアクセサーで参照します。

| 環境変数 | 既定値 | 用途 |
| --- | --- | --- |
| `DEBUG` | `false` | 詳細ログ |
| `LOG_LEVEL` | `INFO` | ログレベル |
| `SAVE_DIRECTORY` | `saves` | セーブディレクトリ |
| `AUTO_SAVE_ENABLED` | `true` | オートセーブ設定 |

設定の実際の値と読み込み順序は `internal/config/config.go` を正とします。ゲームルールや別のゲームモードを設定から切り替える仕組みはありません。
