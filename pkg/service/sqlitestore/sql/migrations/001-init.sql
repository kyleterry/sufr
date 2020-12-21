create table if not exists urls (
  id text not null primary key,
  url text not null unique,
  title text,
  created_at timestamp default CURRENT_TIMESTAMP,
  updated_at timestamp
);

create table if not exists tags (
  id text not null primary key,
  name text not null unique,
  created_at timestamp default CURRENT_TIMESTAMP,
  updated_at timestamp
);

create table if not exists users (
  id text not null primary key,
  email text not null unique,
  password_hash text not null,
  api_token text,
  embed_content boolean not null default false,
  pinned_categories json default '[]',
  created_at timestamp default CURRENT_TIMESTAMP,
  updated_at timestamp
);

create table if not exists user_urls (
  id text not null primary key,
  user_id text not null,
  url_id text not null,
  title text,
  favorite boolean not null default false,
  tags json default '[]',
  created_at timestamp default CURRENT_TIMESTAMP,
  updated_at timestamp,
  foreign key(user_id) references users(id) on delete cascade,
  foreign key(url_id) references urls(id) on delete cascade,
  unique(user_id, url_id)
);

create table if not exists user_url_tags (
  user_url_id text not null,
  tag_id text not null,
  foreign key(user_url_id) references user_urls(id) on delete cascade,
  foreign key(tag_id) references tags(id) on delete cascade,
  unique(user_url_id, tag_id)
);
