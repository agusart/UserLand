CREATE TABLE IF NOT EXISTS users(
    id serial PRIMARY KEY,
    full_name VARCHAR (50) UNIQUE NOT NULL,
    password VARCHAR (50) NOT NULL,
    email VARCHAR (300) UNIQUE NOT NULL,
    verified BOOLEAN default false
);