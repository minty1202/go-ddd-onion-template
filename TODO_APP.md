# Todo App

## 目標

実践的なライブラリを使って Todo アプリの CRUD API を構築する。

## スコープ

- Todo の CRUD
- Todo を Done にする

## 技術スタック

- Go
- DDD / オニオンアーキテクチャ
- Connect RPC
  - **buf** — proto ファイルの管理・lint・コード生成を行う CLI ツール
  - **protoc-gen-go** — .proto から Go のメッセージ型（構造体）を生成する
  - **protoc-gen-connect-go** — .proto から Connect 用の handler / client コードを生成する
  - **connectrpc.com/connect** — Connect RPC のコアライブラリ
  - **connectrpc.com/validate** + **buf.build/go/protovalidate** — proto スキーマに書いたバリデーションルールを自動適用する
- PostgreSQL
  - **sqlc** — SQL から型安全な Go コードを生成する
  - **pgx** — PostgreSQL ドライバ
  - **goose** — マイグレーションツール

## ディレクトリ構造

```
internal/
  domain/ # エンティティ、値オブジェクト、ドメインサービス、リポジトリ interface
  usecase/    # ドメインを組み合わせてアプリケーションの操作を実現する
  infra/ # ドメインが定義した interface の実装
  presentation/  # 外部からのリクエストの入口
```
