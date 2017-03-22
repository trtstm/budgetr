package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/trtstm/budgetr/log"
)

type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

func (v *NullInt64) Set(data int64) {
	v.Int64 = data
	v.Valid = true
}

type NullFloat64 struct {
	sql.NullFloat64
}

func (v NullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullFloat64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *float64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

func (v *NullFloat64) Set(data float64) {
	v.Float64 = data
	v.Valid = true
}

type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

func (v *NullString) Set(data string) {
	v.String = data
	v.Valid = true
}

type categoryResponse struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type expenditureResponse struct {
	ID       int64             `db:"id" json:"id"`
	Amount   float64           `db:"amount" json:"amount"`
	Date     time.Time         `db:"date" json:"date"`
	Category *categoryResponse `json:"category"`
}

func handleError(name string, c echo.Context, err error) (bool, error) {
	if err != nil {
		log.Errorf("%s: %s\n", name, err.Error())
		return false, c.NoContent(http.StatusInternalServerError)
	}

	return true, nil
}

func expenditureCategoriesQuery(tx *sqlx.Tx, ids ...int64) (results map[int64][]*categoryResponse, err error) {
	if len(ids) == 0 {
		return
	}

	query := `
    SELECT
      categories.id as id,
      categories.name as name,
      expenditure_categories.expenditure_id as expenditure_id
    FROM categories
    JOIN expenditure_categories ON expenditure_categories.category_id = categories.id
    WHERE expenditure_categories.expenditure_id in (?)
  `
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return
	}

	query = tx.Rebind(query)

	rows, err := tx.Queryx(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	results = map[int64][]*categoryResponse{}
	for _, id := range ids {
		results[id] = []*categoryResponse{}
	}

	for rows.Next() {
		category := &struct {
			ExpenditureID int64 `db:"expenditure_id"`
			categoryResponse
		}{}
		if err = rows.StructScan(category); err != nil {
			return
		}

		results[category.ExpenditureID] = append(results[category.ExpenditureID], &category.categoryResponse)
	}

	return
}

func assignCategoriesQuery(tx *sqlx.Tx, id int64, categories ...string) (err error) {

	return
}

func indexExpenditure(c echo.Context) error {
	var limit uint64
	var offset uint64

	var err error
	if limit, err = strconv.ParseUint(c.QueryParam("limit"), 10, 64); err != nil {
		limit = 100
	}
	offset, err = strconv.ParseUint(c.QueryParam("offset"), 10, 64)

	var start time.Time
	var end time.Time
	hasDateRange := false

	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")
	if len(startStr) > 0 && len(endStr) > 0 {
		start, _ = time.Parse(time.RFC3339, startStr)
		end, _ = time.Parse(time.RFC3339, endStr)
		hasDateRange = true
	}

	sortColumn := strings.ToLower(strings.TrimSpace(c.QueryParam("sort")))
	sortOrder := strings.ToLower(strings.TrimSpace(c.QueryParam("order")))
	switch sortColumn {
	case "date":
	case "id":
	case "amount":
	default:
		sortColumn = ""
	}

	switch sortOrder {
	case "asc":
	case "desc":
	default:
		sortOrder = "asc"
	}

	if limit > 100 {
		limit = 100
	}

	b := getExpendituresQuery()
	if hasDateRange {
		b = betweenDatesQuery(b, start, end)
	}
	if sortColumn != "" {
		b = sortByQuery(b, sortColumn, sortOrder)
	}
	b = limitQuery(b, limit)
	b = offsetQuery(b, offset)

	q, args, err := b.ToSql()
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not build query.")
		return c.NoContent(http.StatusInternalServerError)
	}

	fmt.Println(q, args)

	rows, err := db.Queryx(q, args...)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not execute query.")
		return c.NoContent(http.StatusInternalServerError)
	}
	defer rows.Close()

	expenditures := []*expenditureResponse{}

	for rows.Next() {
		row := &struct {
			expenditureResponse
			CategoryID   NullInt64  `db:"category_id"`
			CategoryName NullString `db:"category_name"`
		}{}
		if err = rows.StructScan(row); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan result.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if row.CategoryID.Valid {
			row.expenditureResponse.Category = &categoryResponse{
				ID:   row.CategoryID.Int64,
				Name: row.CategoryName.String,
			}
		}

		expenditures = append(expenditures, &row.expenditureResponse)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"offset": offset,
		"limit":  limit,
		"data":   expenditures,
	})
}

func createExpenditure(c echo.Context) error {
	data := &struct {
		Amount   float64   `json:"amount"`
		Date     time.Time `json:"date"`
		Category string    `json:"category"`
	}{}

	if err := c.Bind(data); err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not bind request parameters.")
		return c.NoContent(http.StatusBadRequest)
	}

	expenditure := &expenditureResponse{
		Amount: data.Amount,
		Date:   data.Date,
	}

	columns := []string{"amount", "date"}
	values := []interface{}{expenditure.Amount, expenditure.Date}

	data.Category = strings.TrimSpace(data.Category)
	if len(data.Category) > 0 {
		category := &categoryResponse{
			Name: data.Category,
		}
		if err := createOrGetCategory(category); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not create or get category")
			return c.NoContent(http.StatusInternalServerError)
		}

		columns = append(columns, "category_id")
		values = append(values, category.ID)
		expenditure.Category = category
	}

	b := sq.Insert("expenditures").Columns(columns...).Values(values...)

	q, args, err := b.ToSql()
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not build query.")
		return c.NoContent(http.StatusInternalServerError)
	}

	result, err := db.Exec(q, args...)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could execute query.")
		return c.NoContent(http.StatusInternalServerError)
	}

	expenditure.ID, _ = result.LastInsertId()

	log.Infof("Created expenditure: %+v", expenditure)
	return c.JSON(http.StatusOK, expenditure)
}
