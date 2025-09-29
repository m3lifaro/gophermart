-- +goose Up
create table if not exists users
(
    id       INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login    VARCHAR(10)   not null,
    password VARCHAR(64)   not null,
    balance  float8 not null default 0
);
CREATE UNIQUE INDEX unique_idx_users_login
    ON users (login);

-- +goose Down
drop table if exists users;
DROP INDEX IF EXISTS unique_idx_users_login;