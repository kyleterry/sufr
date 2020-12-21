-- sufr:dialect sqlite3

-- sufr:map_query TagManager.Create
insert or ignore into tags (id, name, created_at, updated_at)
values (:id, :name, :created_at, :updated_at)

-- sufr:map_query TagManager.GetByID
select
  id,
  name,
  created_at,
  updated_at
from tags
where id = ?

-- sufr:map_query TagManager.GetByName
select
  id,
  name,
  created_at,
  updated_at
from tags
where name = ?

-- sufr:map_query URLManager.Create
insert or ignore into urls
  (id, url, title, created_at, updated_at)
values
  (:id, :url, :title, :created_at, :updated_at)

-- sufr:map_query URLManager.GetByURL
select 
  id as id,
  url as url,
  title as title,
  created_at as created_at,
  updated_at as updated_at
from urls
where url = ?

-- sufr:map_query UserManager.Create
insert into users
  (id, email, password_hash, created_at, updated_at)
values
  (:id, :email, :password_hash, :created_at, :updated_at)

-- sufr:map_query UserManager.UpdatePinnedCategories
update users
  set pinned_categories = (
  select
    json_group_array(
      json_object('label', label, 'tags', tags))
  from (
    select
      json_extract(cats.value, '$.label') as label,
      json_group_array(
        json_object('id',
          json_extract(items.value, '$.id'))) as tags
    from json_each(json(?)) cats
    join json_each(cats.value, '$.tags.items') items
    group by json_extract(cats.value, '$.label')))
where id = ?

-- sufr:map_query UserManager.GetByEmail
select
  users.id as id,
  users.email as email,
  users.password_hash as password_hash,
  coalesce(
    nullif(users.api_token, ''), ''
  ) as api_token,
  users.embed_content as embed_content,
  users.created_at as created_at,
  users.updated_at as updated_at
from users
where users.email = ?;

-- sufr:map_query UserManager.GetByID
select
  users.id as id,
  users.email as email,
  users.embed_content as embed_content,
  users.created_at as created_at,
  users.updated_at as updated_at
from users
where users.id = ?;

-- sufr:map_query UserManager.getPinnedCategories
select
  json_extract(cats.value, '$.label') as label,
  json_object('items',
    json_array(
      json_set(jt.value, '$.name', t.name)
    )
  ) as tags
from users
left join json_each(users.pinned_categories) cats
left join json_each(json_extract(cats.value, '$.tags')) jt
left join tags t on t.id = json_extract(jt.value, '$.id')
where users.id = ? and users.pinned_categories != '[]'

-- sufr:map_query UserURLManager.Create
insert into user_urls
  (id, user_id, url_id, title, favorite, created_at, updated_at)
values
  (:id, :user.id, :url.id, :title, :favorite, :created_at, :updated_at)

-- sufr:map_query UserURLManager.Update
update user_urls
  set 
    title = :title,
    favorite = :favorite,
    updated_at = CURRENT_TIMESTAMP
where user_id = :user.id and id = :id

-- sufr:map_query UserURLManager.clearTags
delete from user_url_tags where user_url_id = ?

-- sufr:map_query UserURLManager.updateTags
insert into user_url_tags
  (user_url_id, tag_id)
values
  (?, ?)

-- sufr:map_query UserURLManager.GetAll
select
  uu.id as id,
  u.id as 'url.id',
  u.url as 'url.url',
  u.title as 'url.title',
  uu.user_id as 'user.id',
  uu.title as title,
  coalesce(
    nullif(uu.title, ''),
    u.title
  ) as derived_title,
  json_object('items',
    json_group_array(
      json_object('id', t.id, 'name', t.name)
    )
  ) as tags,
  uu.favorite as favorite,
  uu.created_at,
  uu.updated_at
from
  user_urls uu
join urls u on u.id = uu.url_id
join user_url_tags ut on ut.user_url_id = uu.id
join tags t on t.id = ut.tag_id
where uu.user_id = ?
group by uu.id

-- sufr:map_query UserURLManager.GetAllAfter
select 
    uu.*,
    json_object('items',
      json_group_array(
        json_object('id', t.id, 'name', t.name)
      )
    ) as tags
from (
  select
    row_number() over (
      order by uu.id) as row,
    uu.id as id,
    u.id as 'url.id',
    u.url as 'url.url',
    u.title as 'url.title',
    uu.user_id as 'user.id',
    uu.title as title,
    coalesce(
      nullif(uu.title, ''),
      u.title
    ) as derived_title,
    uu.favorite as favorite,
    uu.created_at as created_at,
    uu.updated_at as updated_at
  from
    user_urls uu
  join urls u on u.id = uu.url_id
  where uu.user_id = ?
) as uu
left join user_url_tags ut on ut.user_url_id = uu.id
left join tags t on t.id = ut.tag_id
where row > ? and row <= ?
group by uu.id

-- sufr:map_query UserURLManager.GetByURLID
select
  uu.id as id,
  u.id as 'url.id',
  u.url as 'url.url',
  u.title as 'url.title',
  uu.user_id as 'user.id',
  uu.title as title,
  coalesce(
    nullif(uu.title, ''),
    u.title
  ) as derived_title,
  json_object('items',
    json_group_array(
      json_object('id', t.id, 'name', t.name)
    )
  ) as tags,
  /* group_concat(t.name) as tags, */
  uu.favorite as favorite,
  uu.created_at,
  uu.updated_at
from
  user_urls uu
join urls u on u.id = uu.url_id
join user_url_tags ut on ut.user_url_id = uu.id
join tags t on t.id = ut.tag_id
where uu.user_id = ? and uu.url_id = ?
group by uu.id

-- sufr:map_query UserURLManager.GetAllByTags
select
  uu.id as id,
  u.id as 'url.id',
  u.url as 'url.url',
  u.title as 'url.title',
  uu.user_id as 'user.id',
  coalesce(nullif(uu.title, ''), u.title) as title,
  uu.favorite as favorite,
  json_object('items',
    json_group_array(
      json_object('id', t.id, 'name', t.name)
    )
  ) as tags,
  uu.created_at,
  uu.updated_at
from
  user_urls uu
join urls u on u.id = uu.url_id
join user_url_tags ut on ut.user_url_id = uu.id
join tags t on t.id = ut.tag_id
where user_id = ? and t.id in(?)
group by uu.id
