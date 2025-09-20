-- +goose Up
create table if not exists users (
    id serial primary key,
    login text not null ,
    password text not null
);
CREATE UNIQUE INDEX unique_idx_users_login
    ON users (login);

-- +goose Down
drop table if exists users;
DROP INDEX IF EXISTS unique_idx_users_login;