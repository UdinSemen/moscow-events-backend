create table public.news_events
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

