create table if not exists session (
    id serial PRIMARY KEY ,
    user_id int not null,
    ip text,
    name text,
    jwt_id text,
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


alter table session add constraint client_session_fkey
    foreign key (user_id)
        references users(id)
        on delete cascade;
