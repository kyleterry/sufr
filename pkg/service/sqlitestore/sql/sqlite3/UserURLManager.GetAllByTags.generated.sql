-- Code generated by build_sql.awk; DO NOT EDIT.
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