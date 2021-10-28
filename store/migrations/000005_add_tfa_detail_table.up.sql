create sequence tfa_detail_seq;
create sequence users_seq;
create sequence sessions_seq;
create sequence client_seq;
create sequence tfa_backup_seq;


create table if not exists tfa_detail(
    id integer primary key,
    user_id integer unique not null ,
    tfa_secret text,
    created_at timestamp,
    deleted_at timestamp,
    activate_at timestamp
);

alter table tfa_detail alter column id set default nextval('tfa_detail_seq');

alter table tfa_detail
    add constraint user_id_fk
    foreign key(user_id)
    references users(id)
    on delete cascade;


alter table users alter column id type integer;
alter table tfa_backup_code alter column id type integer;
alter table session alter column id type integer;
alter table client alter column id type integer;

alter table users alter column id set default nextval('users_seq');
alter table tfa_backup_code alter column id set default nextval('tfa_backup_seq');
alter table session alter column id set default nextval('sessions_seq');
alter table client alter column id set default nextval('client_seq');

select setval('users_seq',  (SELECT MAX(id) FROM users));
select setval('tfa_backup_seq',  (SELECT MAX(id) FROM tfa_backup_code));
select setval('sessions_seq',  (SELECT MAX(id) FROM session));
select setval('client_seq',  (SELECT MAX(id) FROM client));
