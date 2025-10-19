create table if not exists rooms (
  id text primary key,
  created_at timestamptz not null default now()
);

create table if not exists snapshots (
  room_id text not null references rooms(id) on delete cascade,
  seq bigint not null,
  body jsonb not null,
  created_at timestamptz not null default now(),
  primary key (room_id, seq)
);

create table if not exists ops (
  room_id text not null references rooms(id) on delete cascade,
  seq bigint not null,
  body jsonb not null,
  created_at timestamptz not null default now(),
  primary key (room_id, seq)
);
