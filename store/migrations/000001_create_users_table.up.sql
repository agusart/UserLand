CREATE TABLE IF NOT EXISTS users(
    id serial PRIMARY KEY,
    full_name text  NOT NULL,
    password text NOT NULL,
    email text NOT NULL,
    verified BOOLEAN default false,
    tfa_enabled BOOLEAN default false,
    created_at timestamp null,
    deleted_at timestamp null
);