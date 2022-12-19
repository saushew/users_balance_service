CREATE TABLE users (
    id bigserial not null unique,
    balance numeric not null
);

CREATE TABLE transactions (
    id bigserial not null unique,
    user_id bigserial not null,
    amount numeric not null,
    tx_type varchar not null,
    details varchar,
    ts bigserial not null
);