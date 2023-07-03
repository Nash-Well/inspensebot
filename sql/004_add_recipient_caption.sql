-- +goose Up

alter table recipient add column media_caption varchar default '';

-- +goose Down

alter table recipient drop column media_caption;