-- Code generated by build_sql.awk; DO NOT EDIT.
select
  id,
  name,
  created_at,
  updated_at
from tags
where id = ?