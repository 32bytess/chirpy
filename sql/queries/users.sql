-- name: CreateUser :one
insert into
  users (id, created_at, updated_at, email, password)
values
  (gen_random_uuid(), now(), now(), $1, $2)
returning
  *;

-- name: DeleteAllUsers :exec
delete from users;

-- name: GetUserByEmail :one
select
  *
from
  users
where
  email = $1;
