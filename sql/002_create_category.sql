-- +goose Up

create table category(
        id          serial          primary key,
        name        varchar(40)     not null,
        user_id     bigint          not null
);

create table subcategory(
        id              serial          primary key,
        name            varchar(40)     not null,
        category_id     int             not null,

        foreign key (category_id) references category (id)
);