create extension if not exists pgcrypto;

create type content_status as enum ('draft', 'generating', 'ready_to_render', 'rendering', 'ready', 'failed');
create type render_status as enum ('queued', 'rendering', 'done', 'failed');
create type ad_platform as enum ('instagram', 'facebook');

create table content_items (
  id uuid primary key default gen_random_uuid(),
  product text not null,
  title text not null,
  brief text not null,
  raw_video_key text,
  status content_status not null default 'draft',
  caption text,
  script jsonb,
  rendered_video_key text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table render_jobs (
  id uuid primary key default gen_random_uuid(),
  content_item_id uuid not null references content_items(id) on delete cascade,
  status render_status not null default 'queued',
  error text,
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz not null default now()
);

create table ad_entries (
  id uuid primary key default gen_random_uuid(),
  content_item_id uuid references content_items(id) on delete set null,
  platform ad_platform not null,
  external_ad_id text,
  spend numeric(12,2) not null default 0,
  started_at date not null,
  created_at timestamptz not null default now()
);

create table ad_metrics (
  id uuid primary key default gen_random_uuid(),
  ad_entry_id uuid not null references ad_entries(id) on delete cascade,
  date date not null,
  impressions integer not null default 0,
  reach integer not null default 0,
  clicks integer not null default 0,
  spend numeric(12,2) not null default 0,
  created_at timestamptz not null default now(),
  unique (ad_entry_id, date)
);
