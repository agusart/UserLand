create table if not exists password_history (
    id integer primary key,
    user_id integer not null,
    password text not null,
    created_at timestamp
);


create sequence password_history_seq;
alter table password_history alter column id set default nextval('password_history_seq');


alter table password_history
    add constraint user_id_fk
        foreign key(user_id)
            references users(id)
            on delete cascade;