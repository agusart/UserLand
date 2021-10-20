create table if not exists tfa_backup_code (
    id serial primary key,
    code int not null,
    user_id int not null
);

alter table tfa_backup_code
    add constraint tfa_userid_fk
    foreign key (user_id)
    references users (id)
    on delete cascade;