package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/tealeg/xlsx"
	"github.com/trtstm/budgetr/db"
	"github.com/trtstm/budgetr/log"
	"github.com/trtstm/budgetr/models"
)

type exportController struct {
}

func (c *exportController) ExportExcel(ctx echo.Context) error {
	timeStart := time.Now()

	params := []struct {
		Start time.Time `json:"start" form:"start"`
		End   time.Time `json:"end" form:"end"`
		Title string    `json:"title" form:"title"`
	}{}

	var err error
	if len(ctx.FormValue("ranges")) != 0 {
		err = json.Unmarshal([]byte(ctx.FormValue("ranges")), &params)
	} else {
		err = ctx.Bind(&params)
	}

	if err != nil {
		log.Info("ExportController::ExportExcel Failed to bind params: %v", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	results := map[string][]float64{}

	categories := []models.Category{}
	if q := db.DB.Find(&categories); q.Error != nil {
		log.Info("ExportController::ExportExcel Failed to retrieve categories: %v", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	categories = append(categories, models.Category{
		Name: "",
	})

	for _, category := range categories {
		results[category.Name] = make([]float64, len(params))
	}

	for i, r := range params {
		q := categoryStatsQuery()
		q = dateRangeQuery(r.Start, r.End, q)
		stats := []*CategoryStatsResponse{}
		q.Scan(&stats)
		if q.Error != nil {
			log.Errorf("ExportController::ExportExcel Could not execute query: %v", q.Error)
			return ctx.NoContent(http.StatusInternalServerError)
		}

		for _, stat := range stats {
			results[stat.Name.String][i] += stat.Total
		}
	}

	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Uitgaves")
	if err != nil {
		log.Errorf("ExportController::ExportExcel Could not create excel file: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "Categorie"
	for _, param := range params {
		headerRow.AddCell().Value = param.Title
	}

	for category, result := range results {
		row = sheet.AddRow()
		cell = row.AddCell()
		if category == "" {
			category = "geen"
		}
		cell.Value = category

		for _, total := range result {
			cell := row.AddCell()
			cell.SetFloat(total)
		}
	}

	err = file.Save("export.xlsx")
	if err != nil {
		log.Errorf("ExportController::ExportExcel Could not save excel file: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	elapsed := time.Since(timeStart)

	log.Infof("ExportController::ExportExcel Generated excel in %s.", elapsed)
	return ctx.Attachment("export.xlsx", "export.xlsx")
}

// ExportController for /exports endpoint.
var ExportController exportController
