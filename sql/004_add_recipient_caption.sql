-- +goose Up

alter table recipient add column media_caption varchar(128) default '';

-- +goose Down

alter table recipient drop column media_caption;