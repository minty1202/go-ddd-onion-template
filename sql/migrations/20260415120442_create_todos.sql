-- +goose Up
CREATE TABLE todos (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER todos_set_updated_at
BEFORE UPDATE ON todos
FOR EACH ROW
EXECUTE FUNCTION SET_UPDATED_AT();


-- +goose Down
-- pgls-ignore lint/safety/banDropTable
DROP TABLE todos;
