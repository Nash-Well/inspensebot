-- +goose Up

alter table users add column cache json;

-- +goose Down

alter table users drop column cache;