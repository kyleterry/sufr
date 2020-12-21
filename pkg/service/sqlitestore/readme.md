# SUFR sql service

This code implements the `store.Manager` interface for persisting most objects
into a relational database.

## Queries

All queries are defined in [./sql/queries.sql](sql/queries.sql) and annotated
with comments to configure how they map to various manager methods. When a
developer changes a query, they will run `go generate ./...` from the root (or
run it relative to this path) and that in turn will call `awk -f build_sql.awk
queries.sql` which will match the annotations and split the queries into their
respective `.sql` files. Once that's done, another generator runs that embeds
the `sql` directory contents into a go file called `build_assets.go`. This
generated code is a virtual file system that allows us to open a sql file from
Go code like you would open a normal file on the filesystem.

This makes it easy to develop queries inside one file where context about the
data model and db schema is not lost.

### SQL file generation

Lets talk about the [awk script](./build_sql.awk).

Inside you will see some comments that look structured. They follow a `--
sufr:<action> param` pattern. These are matched by awk patterns to tell the
script how to break up the file.

`-- sufr:dialect sqlite` tells the awk script to set the dialect to sqlite and
store generated sql files in `./sql/sqlite`. This can only be specified once and
a warning is returned if it occurs again and setting the dialect is ignored.

`-- sufr:map_query <name>` tells the awk script to set part of the filename.
Name is mostly free-form, but it cannot contain a space. The name part will be
used, along with dialect, to determine where the sql output will go.

Example: 

```sql
-- sufr:dialect sqlite
-- sufr:map_query TagManager.Create
insert into tags (name) values (?)
```

The above example will generate a file called
`sql/sqlite/TagManager.Create.generated.sql`.
