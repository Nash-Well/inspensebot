-- +goose Up

create table finances (
        id                  serial              primary key,
        user_id             bigint              not null,
        type                varchar(30)         not null,
        date                timestamp           not null default now(),
        amount              float               not null default 0,
        category            varchar             not null,
        subcategory         varchar             not null
);

create table recipient (
        id                  serial              primary key,
        finance_id          int                 not null,
        media               varchar(128)        not null default '',
        media_type          varchar(30)         not null default '',
        created_at          timestamp           not null default now(),

        foreign key (finance_id) references finances (id)
);

-- +goose Down

drop table finances;
drop table recipient;