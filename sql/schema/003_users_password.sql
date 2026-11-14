-- +goose Up
alter table users
add password text not null default 'unset';

-- +goose Down
alter table users
drop column password;
