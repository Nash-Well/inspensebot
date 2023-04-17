-- +goose Up

create table users(
        id              bigint           primary key,
        created_at      timestamp        not null default now(),
        language        varchar(2)       not null default 'uk',
        state           varchar(32)      not null default ''
);