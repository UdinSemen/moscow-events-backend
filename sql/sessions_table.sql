create table sessions
(
    id uuid default gen_random_uuid() primary key,
    user_id uuid references users (id),
    refresh_token varchar(1024) not null,
    ip varchar(1024),
    finger_print varchar(2048),
    exp_at timestamp,
    created_at timestamp default now(),
    updated_at timestamp
);