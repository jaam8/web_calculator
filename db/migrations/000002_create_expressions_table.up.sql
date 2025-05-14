create extension if not exists "pgcrypto";

create schema if not exists expressions;

create table if not exists expressions.expressions (
    id uuid not null
     default gen_random_uuid() primary key,
    user_id uuid references users.users(id) on delete cascade,
    status text not null,
    result double precision
);


