create table users (
    id serial primary key,
    telegram_id int not null unique
)