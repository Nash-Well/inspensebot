-- +goose Up


create table finances (
        id                  serial              primary key,
        user_id             bigint              not null,
        type                varchar(30)         not null,
        date                timestamp           not null default now(),
        amount              float               not null default 0,
        category_id         int                 references category (id),
        subcategory_id      int                 references subcategory (category_id),

        foreign key (user_id) references users (id)
);

create table recipient (
        id                  serial              primary key,
        finance_id          int                 not null,
        file_id             varchar             not null,
        media_type          varchar(30)         not null,
        created_at          timestamp           not null default now(),

        foreign key (finance_id) references finances (id)
);
