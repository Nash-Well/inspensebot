-- +goose Up


create table finances (
        id                  serial              primary key,
        user_id             bigint              not null,
        type                varchar(30)         not null,
        date                timestamp           not null default now(),
        amount              float               not null default 0,
        category_id         int                 not null,
        subcategory_id      int                 not null,

        foreign key (user_id) references users (id),
        foreign key (category_id) references category (id),
        foreign key (subcategory_id) references subcategory (id)
);

create table recipient (
        id                  serial              primary key,
        finance_id          int                 not null,
        file_id             varchar             not null,
        media_type          varchar(30)         not null,
        created_at          timestamp           not null default now(),

        foreign key (finance_id) references finances (id)
);
