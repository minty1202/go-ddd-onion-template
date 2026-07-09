-- name: InsertTodo :exec
INSERT INTO todos (
    id, title, body, completed, version
) VALUES (
    $1, $2, $3, $4, $5
);

-- name: UpdateTodo :one
UPDATE todos
SET
    title = $1,
    body = $2,
    completed = $3,
    version = version + 1
WHERE id = $4 AND version = $5
RETURNING version;

-- name: GetTodo :one
SELECT * FROM todos
WHERE id = $1;

-- name: DeleteTodoByID :exec
DELETE FROM todos
WHERE id = $1;
