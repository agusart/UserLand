create table if not exists audit_logs (
    id integer primary key,
    user_id integer not null,
    session_id integer not null,
    remote_ip text,
    created_at timestamp,
    deleted_at timestamp
);


create sequence audit_logs_seq;
alter table audit_logs alter column id set default nextval('audit_logs_seq');


alter table audit_logs
    add constraint log_user_id_fk
        foreign key(user_id)
            references users(id)
            on delete cascade;

alter table audit_logs
    add constraint session_user_id_fk
        foreign key(session_id)
            references session(id)
            on delete cascade;