CREATE TABLE IF NOT EXISTS users(
    id serial PRIMARY KEY,
    full_name text  NOT NULL,
    password text NOT NULL,
    email text UNIQUE NOT NULL,
    verified BOOLEAN default false,
    created_at timestamp null,
    deleted_at timestamp null,
    tfa_enabled BOOLEAN default false
);