create type release_track as enum ('internal', 'closed', 'open', 'production');
create type release_status as enum ('draft', 'uploaded', 'publishing', 'in_review', 'rolled_out', 'halted', 'failed');

create table apps (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  package_name text not null unique,
  play_console_url text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table app_releases (
  id uuid primary key default gen_random_uuid(),
  app_id uuid not null references apps(id) on delete cascade,
  track release_track not null,
  version_code integer not null,
  version_name text,
  release_notes text,
  bundle_key text,
  status release_status not null default 'draft',
  error text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  released_at timestamptz
);

create table app_metrics (
  id uuid primary key default gen_random_uuid(),
  app_id uuid not null references apps(id) on delete cascade,
  date date not null,
  crashes integer,
  anr_count integer,
  rating_avg numeric(3,2),
  rating_count integer,
  created_at timestamptz not null default now(),
  unique (app_id, date)
);
