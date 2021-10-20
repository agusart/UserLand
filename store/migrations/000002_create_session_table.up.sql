create table if not exists session (
    id serial PRIMARY KEY ,
    ip text,
    jwt_id text,
    name text,
    is_current bool default false,
    created_at timestamp,
    deleted_at timestamp
);

create table if not exists client (
    id serial primary key,
    session_id int NOT NULL,
    name text,
    created_at timestamp,
    deleted_at timestamp
);

alter table client add constraint client_session_fkey
    foreign key (session_id)
    references session(id)
    on delete cascade;
