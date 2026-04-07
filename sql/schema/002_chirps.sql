-- +goose up
create table chirps
(
    id         uuid primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    body       text      not null,
    user_id    uuid      not null references users (id) on delete cascade
);

-- +goose down
drop table chirps;
