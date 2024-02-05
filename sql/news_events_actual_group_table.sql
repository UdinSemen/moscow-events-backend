create table public.news_events_actual_group
(
    src      varchar(1024),
    category varchar(1024),
    id_group uuid default gen_random_uuid(),
    PRIMARY KEY (src, category)
);