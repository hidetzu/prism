# ADR-0001: Provider Plugin Architecture for v0.2.0

## Status
Accepted

## Date
2026-04-11

## Context

`prism` v0.2.0 では AWS CodeCommit の Pull Request 取得を追加したい。  
ただし、`prism` 本体の責務は **Pull Request を構造化コンテキストへ分解すること** であり、
各ホスティングサービス固有の API 実装や認証処理を大量に抱え込むことではない。

現状の懸念は次の2点である。

1. **プロバイダー検出の設計**
   - 現在は GitHub 前提で実装されている
   - GitHub / CodeCommit / 将来の Bitbucket などにどう拡張するかを整理する必要がある

2. **認証と依存の違い**
   - GitHub は `GITHUB_TOKEN` を用いた Bearer 認証
   - CodeCommit は AWS credential chain / IAM / region を前提とする
   - Go 本体に AWS SDK を直接組み込むと、依存と責務が膨らむ

また、CodeCommit については既存資産として `ccpr` が存在している。  
この資産を活用しつつ、`prism` 本体の責務を守る必要がある。

---

## Decision

v0.2.0 では、**Provider を外部バイナリプラグイン方式で拡張する**。

`prism` 本体は provider interface と plugin execution layer のみを持ち、
CodeCommit 対応は `ccpr` 系の別バイナリとして実装・連携する。

### 採用する方針

- Provider 拡張は **外部バイナリ方式** を採用する
- `prism-provider-codecommit` のような実行可能バイナリを想定する
- `prism` はサブプロセスとして provider plugin を呼び出す
- provider plugin は PR データを共通 JSON 形式で返す
- `prism` は返却された共通 JSON を `PullRequest` ドメインモデルへ変換して解析を続行する

### GitHub provider の扱い

- v0.2.0 では **GitHub は本体内蔵（built-in）** とする
- `go install` だけで GitHub に対して即座に使えるユーザー体験を維持する
- ただし、内部的には plugin と同等の境界を保ち、将来的に plugin として外出し可能な構造にする
- CodeCommit 以降の外部 provider は plugin 方式とする

### Provider 検出方針

- デフォルトは **URL からの自動判定**
- `--provider` が明示指定された場合は **URL による自動判定を行わない**
- URL 自動判定が難しいケースや GitHub Enterprise などでは `--provider` で明示できる

### 初期ルール

- `github.com` を含む URL は GitHub provider
- CodeCommit URL パターンに一致するものは CodeCommit provider
- `--provider github` などが指定された場合はその provider を強制利用する
- 自動判定できない場合は、分かりやすいエラーメッセージで `--provider` の指定を促す

---

## Rationale

この判断の理由は次の通り。

### 1. `prism` の責務を保てる
`prism` の本質は PR の取得ではなく、**取得済み PR をレビュー用コンテキストへ変換すること** にある。  
プロバイダー固有の実装を本体に抱え込むと、この責務が曖昧になる。

### 2. `ccpr` を活かしやすい
CodeCommit 対応については既存の `ccpr` を活かせる。  
`prism` に AWS SDK を直接入れず、外部 provider としてラップすることで、資産再利用と責務分離を両立できる。

### 3. 将来の拡張に強い
Bitbucket、GitHub Enterprise、GitLab など将来的な provider を想定したとき、
プラグイン方式のほうが `prism` 本体の変更を最小化しやすい。

### 4. 依存分離が明確
GitHub provider は軽量に保てる一方、CodeCommit provider は AWS 認証や SDK を独立して持てる。
これにより本体のビルド、保守、配布がシンプルになる。

---

## Alternatives Considered

## A. 外部バイナリ方式
採用。

### メリット
- 依存が完全に分離できる
- `prism` 本体の責務が明確
- `ccpr` を活かしやすい
- 将来 provider を増やしやすい

### デメリット
- プラグインとの入出力仕様を安定させる必要がある
- サブプロセス実行とエラー処理が必要
- 配布方法を設計する必要がある

## B. Go internal 方式
不採用。

### 理由
- `prism` 本体に provider 固有依存が入りやすい
- AWS SDK などの依存が本体に混ざる
- `prism` の思想である責務分離が崩れやすい

---

## Consequences

### Positive
- v0.2.0 で CodeCommit 対応を追加しつつ、本体をシンプルに保てる
- `ccpr` を provider plugin として再利用する道が開ける
- 将来の provider 拡張戦略が明確になる

### Negative
- plugin protocol を設計する必要がある
- plugin discovery の仕様が必要になる
- ユーザー向けにはインストール説明が少し増える

---

## Plugin Protocol

plugin の呼び出し規約、JSON schema、discovery 仕様の詳細は [Provider Plugin Protocol](../provider-plugin-protocol.md) を参照。

---

## Implementation Status

- [x] plugin protocol の JSON schema を定義する
- [x] provider registry を実装する
- [x] provider auto detection ロジックを実装する
- [x] `--provider` 優先ルールを実装する
- [ ] CodeCommit plugin の第一候補として `ccpr` ラップ方式を検討する
