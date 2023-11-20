-- +goose Up

alter table finances add column location json not null default '';

-- +goose Down

alter table finances drop column location;