package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/trtstm/budgetr/log"
	"github.com/trtstm/budgetr/models"
)

// CategoryStatsResponse contains statistics for a category.
type CategoryStatsResponse struct {
	ID    uint              `json:"id"`
	Name  models.NullString `json:"name"`
	Total float64           `json:"total"`
}

type categoryStatsController struct {
}

func (c *categoryStatsController) Index(ctx echo.Context) error {
	stats := []*CategoryStatsResponse{}

	var start time.Time
	var end time.Time

	if startQ := ctx.QueryParam("start"); len(startQ) > 0 {
		var err error
		start, err = time.Parse(time.RFC3339, startQ)
		if err != nil {
			log.Infof("CategoryStatsController::Index Failed to parse start `%s`: %v", startQ, err)
			return ctx.NoContent(http.StatusBadRequest)
		}
	}

	if endQ := ctx.QueryParam("end"); len(endQ) > 0 {
		var err error
		end, err = time.Parse(time.RFC3339, endQ)
		if err != nil {
			log.Infof("CategoryStatsController::Index Failed to parse end `%s`: %v", endQ, err)
			return ctx.NoContent(http.StatusBadRequest)
		}
	}

	if start.IsZero() != end.IsZero() {
		log.Infof("CategoryStatsController::Index Start and end should both be given.")
		return ctx.NoContent(http.StatusBadRequest)
	}

	q := categoryStatsQuery()
	if !start.IsZero() {
		q = dateRangeQuery(start, end, q)
	}

	q = q.Scan(&stats)

	if q.Error != nil {
		log.Infof("CategoryStatsController::Index Could not execute query: %v", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	logFields := log.Fields{
		"results": len(stats),
	}
	if !start.IsZero() {
		logFields["start"] = start
		logFields["end"] = end
	}

	log.WithFields(logFields).Infof("Returning category statistics.")
	return ctx.JSON(http.StatusOK, stats)
}

// CategoryStatsController for /stats/category endpoint.
var CategoryStatsController categoryStatsController
