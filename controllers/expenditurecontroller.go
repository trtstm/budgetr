package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/trtstm/budgetr/db"
	"github.com/trtstm/budgetr/log"
	"github.com/trtstm/budgetr/models"
)

type colSort struct {
	col   string
	order string
}

func parseSortParam(param string, validCols ...string) *colSort {
	parts := strings.Split(param, "|")
	if len(parts) == 0 {
		return nil
	}

	col := strings.ToLower(parts[0])
	valid := false
	for _, c := range validCols {
		if c == strings.ToLower(col) {
			valid = true
			break
		}
	}
	if !valid {
		return nil
	}

	order := "asc"
	if len(parts) > 1 {
		parts[1] = strings.ToLower(parts[1])
		switch parts[1] {
		case "asc":
			fallthrough
		case "desc":
			order = parts[1]
		}
	}

	return &colSort{col: col, order: order}
}

func sortQuery(cs *colSort, q *gorm.DB) *gorm.DB {
	if cs == nil {
		return q
	}

	return q.Order(cs.col + " " + cs.order)
}

func limitQuery(limit uint, q *gorm.DB) (uint, *gorm.DB) {
	n := limit
	if n == 0 {
		n = 100
	}

	if n > 100 {
		n = 100
	}

	return uint(n), q.Limit(uint(n))
}

func offsetQuery(offset uint, q *gorm.DB) (uint, *gorm.DB) {
	n := offset

	return uint(n), q.Offset(uint(n))
}

func dateRangeQuery(start time.Time, end time.Time, q *gorm.DB) *gorm.DB {
	return q.Where("date >= ? AND date < ?", start, end)
}

type expenditureController struct {
}

func (c *expenditureController) Index(ctx echo.Context) error {
	expenditures := []*models.Expenditure{}

	params := &struct {
		Limit  uint `json:"limit" form:"limit" query:"limit"`
		Offset uint `json:"offset" form:"offset" query:"offset"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Infof("ExpenditureController::Index Could not bind params: '%v'.", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	var limit uint
	var offset uint

	q := db.DB.Preload("Category")

	var start time.Time
	var end time.Time
	if len(ctx.QueryParam("start")) > 0 && len(ctx.QueryParam("end")) > 0 {
		var err error
		start, err = time.Parse(time.RFC3339, ctx.QueryParam("start"))
		if err != nil {
			log.Infof("ExpenditureController::Index Could not parse start: '%v'.", err)
		}
		end, err = time.Parse(time.RFC3339, ctx.QueryParam("end"))
		if err != nil {
			log.Infof("ExpenditureController::Index Could not parse end: '%v'.", err)
		}

		if err == nil {
			q = dateRangeQuery(start, end, q)
		}
	}

	q = sortQuery(parseSortParam(ctx.QueryParam("sort"), "id", "amount", "date"), q)

	limit, q = limitQuery(params.Limit, q)
	offset, q = offsetQuery(params.Offset, q)
	q.Find(&expenditures)
	if q.Error != nil {
		log.Errorf("ExpenditureController::Index Failed to execute query: %v", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.WithFields(log.Fields{"limit": limit, "offset": offset, "size": len(expenditures)}).Infof("ExpenditureController::Index Returning expenditure index.")
	return ctx.JSON(http.StatusOK, echo.Map{
		"data":   TransformExpenditure(expenditures...),
		"limit":  limit,
		"offset": offset,
	})
}

func (c *expenditureController) Show(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		log.Infof("ExpenditureController::Show Could not parse id `%d`: '%v'.", id, err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	expenditure := &models.Expenditure{}
	q := db.DB.Preload("Category").Where("id = ?", id).First(expenditure)
	if q.RecordNotFound() {
		log.Infof("ExpenditureController::Show Expenditure '%d' not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	}

	log.Infof("ExpenditureController::Show Returning expenditure: %+v.", expenditure)
	return ctx.JSON(http.StatusOK, TransformExpenditure(expenditure)[0])
}

func (c *expenditureController) Create(ctx echo.Context) error {
	expenditure := &models.Expenditure{}
	params := &struct {
		Date     time.Time `json:"date" form:"date"`
		Amount   float64   `json:"amount" form:"amount"`
		Category string    `json:"category" form:"category"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Infof("ExpenditureController::Create Could not bind params: '%v'.", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	var category *models.Category

	params.Category = strings.TrimSpace(params.Category)
	if len(params.Category) != 0 {
		category = &models.Category{Name: params.Category}
		if q := db.DB.FirstOrCreate(category, "name = ?", category.Name); q.Error != nil {
			log.Errorf("ExpenditureController::Create FirstOrCreate failed: '%v'.", q.Error)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	expenditure.Amount = params.Amount
	expenditure.Date = params.Date
	expenditure.Category = category

	if q := db.DB.Create(expenditure); q.Error != nil {
		log.Errorf("ExpenditureController::Create Create failed: '%v'.", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.Infof("ExpenditureController::Create Expenditure created: %+v.", expenditure)
	return ctx.JSON(http.StatusCreated, TransformExpenditure(expenditure)[0])
}

func (c *expenditureController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		log.Infof("ExpenditureController::Update Could not parse id `%d`: '%v'.", id, err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	expenditure := &models.Expenditure{}
	if q := db.DB.First(expenditure, "id = ?", id); q.Error != nil {
		if q.RecordNotFound() {
			log.Infof("ExpenditureController::Update Expenditure '%d' not found.", id)
			return ctx.NoContent(http.StatusNotFound)
		}

		log.Errorf("ExpenditureController::Update First failed: '%v'.", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	params := &struct {
		Date     time.Time `json:"date" form:"date"`
		Amount   float64   `json:"amount" form:"amount"`
		Category string    `json:"category" form:"category"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Infof("ExpenditureController::Update Could not bind params: '%v'.", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	var category *models.Category

	params.Category = strings.TrimSpace(params.Category)
	if len(params.Category) != 0 {
		category = &models.Category{Name: params.Category}
		if q := db.DB.FirstOrCreate(category, "name = ?", category.Name); q.Error != nil {
			log.Errorf("ExpenditureController::Update FirstOrCreate failed: '%v'.", q.Error)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	expenditure.Amount = params.Amount
	expenditure.Date = params.Date
	expenditure.Category = category

	q := db.DB.Save(expenditure)
	if q.Error != nil {
		log.Errorf("ExpenditureController::Update Update failed: '%v'.", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if q.RowsAffected == 0 {
		log.Infof("ExpenditureController::Update No rows updated")
		return ctx.NoContent(http.StatusNotFound)
	}

	log.Infof("ExpenditureController::Update Expenditure updated: %+v.", expenditure)
	return ctx.JSON(http.StatusOK, TransformExpenditure(expenditure)[0])
}

func (c *expenditureController) Delete(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		log.Infof("ExpenditureController::Delete Could not parse id `%d`: '%v'.", id, err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	q := db.DB.Where("id = ?", id).Delete(&models.Expenditure{})
	if q.Error != nil {
		log.Errorf("ExpenditureController::Delete Delete failed: '%v'.", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if q.RowsAffected == 0 {
		log.Infof("ExpenditureController::Delete Could not delete expenditure `%d`. Does not exist.", id)
		return ctx.NoContent(http.StatusNotFound)
	}

	log.Infof("ExpenditureController::Delete Expenditure '%d' deleted.", id)
	return ctx.NoContent(http.StatusOK)
}

// ExpenditureController Contains the actions for the 'expenditures' endpoint.
var ExpenditureController expenditureController
