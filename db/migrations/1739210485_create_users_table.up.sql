create table users (
    id serial primary key,
    telegram_id int not null unique,
    created_at timestamp not null default (timezone('utc', now()))
)