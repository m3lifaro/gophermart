-- +goose Up
create table if not exists withdrawals (
    user_id bigint not null ,
    order_id text not null ,
    processed_at timestamp not null default now(),
    amount float8
);

CREATE UNIQUE INDEX unique_idx_withdrawals_order_id
    ON withdrawals (order_id);

-- +goose Down
drop table if exists withdrawals;
DROP INDEX IF EXISTS unique_idx_withdrawals_order_id;