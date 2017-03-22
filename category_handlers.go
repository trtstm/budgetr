package main

import (
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func createOrGetCategory(category *categoryResponse) error {
	category.Name = strings.TrimSpace(category.Name)
	if len(category.Name) == 0 {
		return errors.New("Category name can't be empty.")
	}

	q, args, err := sq.Select("id", "name").From("categories").Where("name = ?", category.Name).ToSql()
	if err != nil {
		return err
	}

	if err = db.QueryRowx(q, args...).StructScan(category); err != nil && err != sql.ErrNoRows {
		return err
	} else if err == nil {
		return nil
	}

	q, args, err = sq.Insert("categories").Columns("name").Values(category.Name).ToSql()
	if err != nil {
		return err
	}

	result, err := db.Exec(q, args...)
	if err != nil {
		return err
	}

	category.ID, _ = result.LastInsertId()

	return nil
}
