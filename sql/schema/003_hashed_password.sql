-- +goose up
alter table users
    add column hashed_password text default 'unset' not null;

-- +goose down
alter table users
    drop column hashed_password;
