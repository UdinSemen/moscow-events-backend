create table public.favourite_list
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