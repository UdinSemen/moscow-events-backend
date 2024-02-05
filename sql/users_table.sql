create table users
(
    id uuid default gen_random_uuid() primary key,
    tg_user_id bigint unique,
    first_name varchar(1024),
    last_name varchar(1024),
    sex varchar(1024),
    role varchar(1024) references roles(type),
    created_at timestamp default now(),
    updated_at timestamp
);