# 現行状態遷移

状態名は `internal/core/state/state.go` の `GameState`、実際の遷移は各画面の `HandleInput` と `internal/core/engine.go` にあります。以下は現在の画面状態とゲーム終了境界のみを示します。

```mermaid
stateDiagram-v2
    [*] --> Menu
    Menu --> Game: start new game or load
    Game --> Help: open help
    Help --> Game: return
    Game --> Symbol: explain symbol
    Symbol --> Game: return
    Game --> SaveLoad: open save/load
    SaveLoad --> Game: resume or restore
    Game --> Victory: return to surface with amulet
    Game --> GameOver: player dies
    Victory --> [*]
    GameOver --> Game: start a fresh game
    GameOver --> [*]: quit
```

## 行動のターン順

GUI と CLI の双方は同じ `internal/core/command.Executor` を呼びます。選択入力などでターンを消費しない場合はモンスター更新を実行しません。ターンを消費する行動ではプレイヤー行動の適用後にモンスター側の更新を行います。細部は executor とパッケージテストが根拠です。
