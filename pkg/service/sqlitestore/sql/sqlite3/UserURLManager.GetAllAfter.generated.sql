-- Code generated by build_sql.awk; DO NOT EDIT.
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
