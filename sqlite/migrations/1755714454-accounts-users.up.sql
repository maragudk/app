create table accounts (
  id text primary key default ('a_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  name text not null
) strict;

create trigger accounts_updated_timestamp after update on accounts begin
  update accounts set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create table users (
  id text primary key default ('u_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  account_id text not null references accounts (id) on delete cascade,
  name text not null,
  email text unique not null,
  confirmed int not null default 0 check ( confirmed in (0, 1) ),
  active int not null default 1 check ( active in (0, 1) )
) strict;

create trigger users_updated_timestamp after update on users begin
  update users set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;
