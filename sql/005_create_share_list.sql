-- +goose Up

create table share_list (
        id                  serial          primary key,
        from_user           bigint          not null,
        forward_from        bigint          not null,
        from_name           varchar(35)     not null default '',
        share_type          varchar(10)     not null default '',
        created_at          timestamp       not null default now(),

        foreign key (from_user) references users (id),
        foreign key (forward_from) references users (id)
);

-- +goose Down

drop table share_list;