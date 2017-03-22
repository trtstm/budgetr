package main

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

func getExpendituresQuery() sq.SelectBuilder {
	query := sq.Select(
		"expenditures.id as id", "expenditures.amount as amount", "expenditures.date as date",
		"categories.id as category_id",
		"categories.name as category_name",
	).From("expenditures").LeftJoin("categories ON expenditures.category_id = categories.id")

	return query
}

func betweenDatesQuery(q sq.SelectBuilder, start, end time.Time) sq.SelectBuilder {
	return q.Where("date >= ? AND date < ?", start, end)
}

func limitQuery(q sq.SelectBuilder, limit uint64) sq.SelectBuilder {
	return q.Limit(limit)
}

func offsetQuery(q sq.SelectBuilder, offset uint64) sq.SelectBuilder {
	return q.Offset(offset)
}

func sortByQuery(q sq.SelectBuilder, column string, order string) sq.SelectBuilder {
	return q.OrderBy(column + " " + order)
}
