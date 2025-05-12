create extension if not exists "pgcrypto";

create schema if not exists users;

create table if not exists users.users (
   id uuid not null
      default gen_random_uuid() primary key,
   login varchar(40) not null unique,
   password_hash varchar(255) not null,
   created_at timestamp default now()
);
