create table if not exists public.news_events
(
    id          uuid primary key default gen_random_uuid(),
    id_group    uuid,          -- уникальный uuid обновления (загрузки)
    src         varchar(1024), -- источник
    category    varchar(1024),
    label       text,
    description text,
    price       text,
    url         varchar(1024),
    url_img     varchar(1024),
    url_buy     varchar(1024),
    created_at  timestamp default now(),
    updated_at  timestamp
);

create table if not exists public.news_events_actual_group
(
    src      varchar(1024),
    category varchar(1024),
    id_group uuid default gen_random_uuid(),
    PRIMARY KEY (src, category)
);

create table if not exists public.dates
(
    id       uuid primary key default gen_random_uuid(),
    id_event uuid references public.news_events (id),
    date     date
);

create table if not exists roles
(
    id uuid default gen_random_uuid() primary key,
    type varchar(1024) unique,
    created_at timestamp default now(),
    updated_at timestamp
);

create table if not exists users
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

create table if not exists sessions
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

create table if not exists public.favourite_list
(
    id         uuid      default gen_random_uuid(),
    user_id uuid references users(id),
    user_tg_id    bigint references users(tg_user_id),
    id_event   uuid,
    id_date    uuid,
    created_at timestamp default now(),
    foreign key (id_event) references news_events (id),
    foreign key (id_date) references dates (id)
);



