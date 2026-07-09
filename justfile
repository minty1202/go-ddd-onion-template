set dotenv-load := true

# プロジェクトのセットアップ（依存のダウンロード）
setup:
    go mod download
    go tool buf dep update

# 静的解析（golangci-lint + sqlc compile + buf lint）
check target="./...":
    golangci-lint run {{target}}
    go tool sqlc compile
    go tool buf lint

# コードフォーマット（goimports + gofumpt + buf format）
fmt target=".":
    goimports -w {{target}}
    gofumpt -w {{target}}
    go tool buf format -w

# テスト実行
test target="./...":
    go test {{target}}

# 依存の整理（不要な依存の削除、不足の追加）
deps:
    go mod tidy
    go tool buf dep update

# DB 起動
db-up:
    docker compose up -d

# DB 停止
db-down:
    docker compose down

# DB 停止 + データ削除
db-reset:
    docker compose down -v

# DB に psql で接続
psql:
    docker compose exec db psql -U ${POSTGRES_USER} -d ${POSTGRES_DB}

# マイグレーション（up|down|reset|create <name>|status）
migrate *args:
    ./scripts/migrate.sh {{args}}

# モック生成（毎回クリーン再生成）
gen-mock:
    rm -rf mocks
    go tool mockery

# sqlc コード生成
gen-sqlc:
    go tool sqlc generate

# proto から Go コード生成（Connect RPC）
gen-proto:
    go tool buf generate

# 開発サーバ起動（air で hot reload）
# APP_ENV=development が必須。direnv で .envrc を読み込むか、明示的に指定すること
dev:
    @if [ "$APP_ENV" != "development" ]; then \
        echo "ERROR: APP_ENV must be 'development'"; \
        echo "Usage: enable direnv (cp .envrc.example .envrc && direnv allow .) or 'APP_ENV=development just dev'"; \
        exit 1; \
    fi
    go tool air

# Connect RPC で Define を叩く（疎通確認用、-v でレスポンスヘッダ表示）
curl-define:
    go tool buf curl -v \
        --schema proto/todo/v1/todo.proto \
        --data '{"title":"test","body":"hello"}' \
        http://localhost:8080/todo.v1.TodoService/Define

# Connect RPC で View を叩く（ID は実際の ULID に書き換える、-v でレスポンスヘッダ表示）
curl-view id:
    go tool buf curl -v \
        --schema proto/todo/v1/todo.proto \
        --data '{"id":"{{id}}"}' \
        http://localhost:8080/todo.v1.TodoService/View
