create table public.dates
(
    id       uuid primary key default gen_random_uuid(),
    id_event uuid references public.news_events (id),
    date     date
);

