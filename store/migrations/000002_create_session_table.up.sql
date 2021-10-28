create table if not exists session (
    id serial PRIMARY KEY ,
    user_id int not null,
    ip text,
    client_id int,
    jwt_id text,
    created_at timestamp,
    deleted_at timestamp
);

create table if not exists client (
    id serial primary key,
    name text unique,
    created_at timestamp,
    deleted_at timestamp
);


alter table session add constraint users_session_fkey
    foreign key (user_id)
        references users(id)
        on delete cascade;


alter table session add constraint client_session_fkey
    foreign key (client_id)
        references client(id)
        on delete cascade;