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

-- name: UpdateUser :one
update users
set
  email = $2,
  password = $3,
  updated_at = now()
where
  id = $1
returning
  *;

-- name: UpgradeUserToChirpyRed :execrows
update users
set
  is_chirpy_red = true,
  updated_at = now()
where
  id = $1;
