package postgres

import (
	"fmt"
	"log"
	"strings"

	"github.com/toventang/debezium-client/adapter"
	"github.com/toventang/debezium-client/schema"
)

func prepareInsertSQL(row schema.Row) string {
	var builder strings.Builder
	var fields strings.Builder
	var values strings.Builder

	l := len(row.FieldItems)
	for i, f := range row.FieldItems {
		fields.WriteString(`"`)
		fields.WriteString(f.Field)
		fields.WriteString(`"`)

		v := adapter.GetValue(f)
		values.WriteString(v)
		if i < l-1 {
			fields.WriteString(",")
			values.WriteString(",")
		}
	}

	tn := fmt.Sprintf(`"%s"."%s"`, row.Schema, row.TableName)
	builder.Grow(len(tn) + len(fields.String()) + len(values.String()) + 23)
	builder.WriteString("INSERT INTO ")
	builder.WriteString(tn)
	builder.WriteString("(")
	builder.WriteString(fields.String())
	builder.WriteString(")VALUES(")
	builder.WriteString(values.String())
	builder.WriteString(")")

	log.Println("insert sql: ", builder.String())
	return builder.String()
}

func prepareUpdateSQL(row schema.Row) string {
	var builder, values, where strings.Builder
	hasCond := false

	l := len(row.FieldItems)
	for i, f := range row.FieldItems {
		v := adapter.GetValue(f)
		if f.PrimaryKey {
			if hasCond {
				where.WriteString(" AND ")
			}
			hasCond = true
			where.WriteString(`"`)
			where.WriteString(f.Field)
			where.WriteString(`"=`)
			where.WriteString(v)
		} else {
			values.WriteString(`"`)
			values.WriteString(f.Field)
			values.WriteString(`"=`)
			values.WriteString(v)

			if i < l-1 {
				values.WriteString(",")
			}
		}
	}

	tn := fmt.Sprintf(`"%s"."%s"`, row.Schema, row.TableName)
	builder.Grow(len(tn) + len(values.String()) + len(where.String()) + 20)
	builder.WriteString("UPDATE ")
	builder.WriteString(tn)
	builder.WriteString(" SET ")
	builder.WriteString(values.String())

	w := where.String()
	if len(w) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(w)
	}

	log.Println("update sql: ", builder.String())
	return builder.String()
}

func prepareUpsertSQL(row schema.Row) string {
	insertSQL := prepareInsertSQL(row)
	var upsertSQL strings.Builder
	var pk string
	l := len(row.FieldItems)
	upsertSQL.WriteString(" UPDATE SET ")
	for i, f := range row.FieldItems {
		v := adapter.GetValue(f)
		if f.PrimaryKey && len(pk) == 0 {
			pk = f.Field
			continue
		}
		upsertSQL.WriteString(`"`)
		upsertSQL.WriteString(f.Field)
		upsertSQL.WriteString(`"=`)
		upsertSQL.WriteString(v)

		if i < l-1 {
			upsertSQL.WriteString(",")
		}
	}
	sql := fmt.Sprintf(`%s ON CONFLICT("%s") DO %s`, insertSQL, pk, upsertSQL.String())

	log.Println("upsert sql: ", sql)
	return sql
}
