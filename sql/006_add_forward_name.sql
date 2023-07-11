-- +goose Up

alter table share_list add column forward_name varchar(35) not null default '';

-- +goose Down

alter table share_list drop column forward_name;