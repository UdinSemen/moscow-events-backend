create table roles
(
    id uuid default gen_random_uuid() primary key,
    type varchar(1024) unique,
    created_at timestamp default now(),
    updated_at timestamp
);