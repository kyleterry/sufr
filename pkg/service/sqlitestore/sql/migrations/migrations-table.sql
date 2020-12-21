create table if not exists migrations (
  version text primary key,
  created_at timestamp not null default CURRENT_TIMESTAMP
);
