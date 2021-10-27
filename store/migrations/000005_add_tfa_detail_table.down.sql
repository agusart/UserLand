alter table tfa_detail alter column id drop default;
alter table users alter column id drop default;
alter table tfa_backup_code alter column id drop default;
alter table session alter column id drop default;
alter table client alter column id drop default;


drop sequence users_seq;
drop sequence sessions_seq;
drop sequence client_seq;
drop sequence tfa_backup_seq;

drop table if exists tfa_detail;
drop sequence tfa_detail_seq;
