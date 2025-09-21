-- +goose Up
create table if not exists user_orders (
    user_id bigint not null,
    order_id text not null ,
    added_at timestamp not null default now(),
    accrual float8,
    status text not null default 'NEW'
);

CREATE UNIQUE INDEX unique_idx_user_orders_order_id
    ON user_orders (order_id);

-- +goose Down
drop table if exists user_orders;
DROP INDEX IF EXISTS unique_idx_user_orders_order_id;