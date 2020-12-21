//go:generate awk -f build_sql.awk sql/queries.sql
//go:generate go run build_tool.go

package sqlitestore
